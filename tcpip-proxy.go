package main

import (
  "encoding/hex"
  "flag"
  "fmt"
  "net"
  "os"
  "runtime"
  "strings"
  "time"
  "logger"
)

var (
  host *string = flag.String("host", "", "target host or address")
  port *string = flag.String("port", "0", "target port")
  listenPort *string = flag.String("listenPort", "0", "listen port")
)

func warn(format string, v ...interface{}) {
  os.Stderr.WriteString(fmt.Sprintf(format + "\n", v...))
}

func die(format string, v ...interface{}) {
  warn(format, v...)
  os.Exit(1)
}


func printableAddr(a net.Addr) string {
  return strings.Replace(a.String(), ":", "-", -1)
}

const (
  LocalToRemote = iota
  RemoteToLocal
)

type Channel struct {
  from,   to            net.Conn
  connectionLogger, binaryLogger *logger.Logger
  ack                   chan bool
}

func NewChannel(from, to net.Conn, peer string, connectionNumber int, connectionLogger *logger.Logger, ack chan bool) *Channel {
  binaryLogger := logger.NewBinaryLogger(connectionNumber, peer)
  channel := &Channel{ from: from, to: to, connectionLogger: connectionLogger, binaryLogger: binaryLogger, ack: ack }

  go channel.PassThrough()
  return channel
}

func (channel Channel) Log(format string, v ...interface{}) {
  channel.connectionLogger.Log(format, v...)
}

func (channel Channel) LogHex(bytes []byte) {
  channel.Log(hex.Dump(bytes))
}

func (channel Channel) LogBinary(bytes []byte) {
  channel.binaryLogger.LogBinary(bytes)
}

func (channel Channel) Read(buffer []byte) (n int, err error) {
  return channel.from.Read(buffer)
}

func (channel Channel) Write(buffer []byte) (n int, err error) {
  return channel.to.Write(buffer)
}

func (channel Channel) Disconnect() {
  channel.Log("Disconnected from %s", channel.FromAddr())
  channel.from.Close()
  channel.to.Close()
  channel.binaryLogger.Close()
  channel.ack <- true
}

func (channel Channel) FromAddr() (addr string) {
  return printableAddr(channel.from.LocalAddr())
}

func (channel Channel) ToAddr() (addr string) {
  return printableAddr(channel.to.LocalAddr())
}

func (channel Channel) PassThrough() {
  b := make([]byte, 10240)
  offset := 0
  packetNumber := 0

  for {
    n, err := channel.Read(b)
    if err != nil {
      break
    }

    if n <= 0 {
      continue
    }

    channel.Log("Received (#%d, %08X) %d bytes from %s", packetNumber, offset, n, channel.FromAddr())

    channel.LogHex(b[:n])
    channel.LogBinary(b[:n])

    channel.Write(b[:n])

    channel.Log("Sent (#%d) to %s\n", packetNumber, channel.ToAddr())

    offset += n
    packetNumber += 1
  }

  channel.Disconnect()
}

type Connection struct {
  local, remote net.Conn
  connectionNumber int
  target string
  logger *logger.Logger
  ack chan bool
}

func NewConnection(local net.Conn, connectionNumber int, target string) *Connection {
  remote, err := net.Dial("tcp", target)
  if err != nil {
    die("Unable to connect to %s, %v", target, err)
  }

  connection := &Connection{ local: local, remote: remote, connectionNumber: connectionNumber, target: target, ack: make(chan bool) }
  go connection.Process()
  return connection
}

func (connection Connection) LocalInfo() string {
  return printableAddr(connection.remote.LocalAddr())
}

func (connection Connection) RemoteInfo() string {
  return printableAddr(connection.remote.RemoteAddr())
}

func (connection Connection) Info(direction int) string {
  switch direction {
  case LocalToRemote:
    return connection.LocalInfo()
  case RemoteToLocal:
    return connection.RemoteInfo()
  }

  panic("Unreachable.")
}

func (connection Connection) From(direction int) net.Conn {
  switch direction {
  case LocalToRemote:
    return connection.local
  case RemoteToLocal:
    return connection.remote
  }

  panic("Unreachable.")
}

func (connection Connection) To(direction int) net.Conn {
  switch direction {
  case LocalToRemote:
    return connection.remote
  case RemoteToLocal:
    return connection.local
  }

  panic("Unreachable.")
}

func (connection Connection) NewChannel(direction int, connectionLogger *logger.Logger) *Channel {
  return NewChannel(connection.From(direction), connection.To(direction), connection.Info(direction), connection.connectionNumber, connectionLogger, connection.ack)
}

func (connection Connection) Process() {
  connectionLogger := logger.NewConnectionLogger(connection.connectionNumber, connection.LocalInfo(), connection.RemoteInfo())
  defer connectionLogger.Close()

  started := time.Now()

  connectionLogger.Log("Connected to %s.\n", connection.target)

  connection.NewChannel(LocalToRemote, connectionLogger)
  connection.NewChannel(RemoteToLocal, connectionLogger)

  // Wait for acks from *both* the pass through channels.
  <-connection.ack
  <-connection.ack

  finished := time.Now()
  duration := finished.Sub(started)

  connectionLogger.Log("Disconnected from %s, duration %s.\n", connection.target, duration.String())
}

type Proxy struct {
  target string
  localPort string
  connectionNumber int
}

func RunProxy(targetHost, targetPort, localPort string) {
  target := net.JoinHostPort(targetHost, targetPort)
  proxy := &Proxy{ target: target, localPort: localPort, connectionNumber: 1 }
  proxy.Run()
}

func (proxy Proxy) Run() {
  fmt.Printf("Start listening on port %s and forwarding data to %s\n", proxy.localPort, proxy.target)

  ln, err := net.Listen("tcp", ":" + proxy.localPort)
  if err != nil {
    die("Unable to start listener %v", err)
  }

  for {
    conn, err := ln.Accept()
    if err != nil {
      warn("Accept failed: %v", err)
      continue
    }

    proxy.ProcessConnection(conn)
  }
}

func (proxy Proxy) ProcessConnection(connection net.Conn) {
  NewConnection(connection, proxy.connectionNumber, proxy.target)

  proxy.connectionNumber += 1
}

func main() {
  runtime.GOMAXPROCS(runtime.NumCPU())
  flag.Parse()
  if flag.NFlag() != 3 {
    warn("Usage: tcpip-proxy -host targetHost -port targetPort -listenPort localPort")
    flag.PrintDefaults()
    os.Exit(1)
  }

  RunProxy(*host, *port, *listenPort)
}
