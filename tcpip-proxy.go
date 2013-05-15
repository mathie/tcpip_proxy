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
  listen_port *string = flag.String("listen_port", "0", "listen port")
)

func warn(format string, v ...interface{}) {
  os.Stderr.WriteString(fmt.Sprintf(format + "\n", v...))
}

func die(format string, v ...interface{}) {
  warn(format, v...)
  os.Exit(1)
}


func format_time(t time.Time) string {
  return t.Format("2006.01.02-15.04.05")
}

func printable_addr(a net.Addr) string {
  return strings.Replace(a.String(), ":", "-", -1)
}

const (
  LocalToRemote = iota
  RemoteToLocal
)

type Channel struct {
  from,   to            net.Conn
  connectionLogger, binary_logger *logger.Logger
  ack                   chan bool
}

func NewChannel(connection *Connection, direction int, connectionLogger *logger.Logger) *Channel {
  peer := connection.Info(direction)
  binaryLogger := logger.NewBinaryLogger(connection.connectionNumber, peer)
  channel := &Channel{ from: connection.From(direction), to: connection.To(direction), connectionLogger: connectionLogger, binary_logger: binaryLogger, ack: connection.ack }

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
  channel.binary_logger.LogBinary(bytes)
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
  channel.binary_logger.Close()
  channel.ack <- true
}

func (channel Channel) FromAddr() (addr string) {
  return printable_addr(channel.from.LocalAddr())
}

func (channel Channel) ToAddr() (addr string) {
  return printable_addr(channel.to.LocalAddr())
}

func (channel Channel) PassThrough() {
  b := make([]byte, 10240)
  offset := 0
  packet_n := 0

  for {
    n, err := channel.Read(b)
    if err != nil {
      break
    }

    if n <= 0 {
      continue
    }

    channel.Log("Received (#%d, %08X) %d bytes from %s", packet_n, offset, n, channel.FromAddr())

    channel.LogHex(b[:n])
    channel.LogBinary(b[:n])

    channel.Write(b[:n])

    channel.Log("Sent (#%d) to %s\n", packet_n, channel.ToAddr())

    offset += n
    packet_n += 1
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
  return printable_addr(connection.remote.LocalAddr())
}

func (connection Connection) RemoteInfo() string {
  return printable_addr(connection.remote.RemoteAddr())
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

func (connection Connection) Process() {
  connectionLogger := logger.NewConnectionLogger(connection.connectionNumber, connection.LocalInfo(), connection.RemoteInfo())
  defer connectionLogger.Close()

  started := time.Now()

  connectionLogger.Log("Connected to %s at %s\n", connection.target, format_time(started))

  NewChannel(&connection, LocalToRemote, connectionLogger)
  NewChannel(&connection, RemoteToLocal, connectionLogger)

  // Wait for acks from *both* the pass through channels.
  <-connection.ack
  <-connection.ack

  finished := time.Now()
  duration := finished.Sub(started)

  connectionLogger.Log("Disconnected from %s at %s, duration %s\n", connection.target, format_time(finished), duration.String())
}

type Proxy struct {
  target string
  local_port string
  conn_n int
}

func RunProxy(target_host, target_port, local_port string) {
  target := net.JoinHostPort(target_host, target_port)
  proxy := &Proxy{ target: target, local_port: local_port, conn_n: 1 }
  proxy.Run()
}

func (proxy Proxy) Run() {
  fmt.Printf("Start listening on port %s and forwarding data to %s\n", proxy.local_port, proxy.target)

  ln, err := net.Listen("tcp", ":" + proxy.local_port)
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
  NewConnection(connection, proxy.conn_n, proxy.target)

  proxy.conn_n += 1
}

func main() {
  runtime.GOMAXPROCS(runtime.NumCPU())
  flag.Parse()
  if flag.NFlag() != 3 {
    warn("Usage: tcpip-proxy -host target_host -port target_port -listen_port local_port")
    flag.PrintDefaults()
    os.Exit(1)
  }

  RunProxy(*host, *port, *listen_port)
}
