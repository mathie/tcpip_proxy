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

type Logger struct {
  filename string
  data chan []byte
}

func NewLogger(filename string) *Logger {
  logger := &Logger { data: make(chan []byte), filename: filename }
  go logger.LoggerLoop()
  return logger
}

func ConnectionLoggerFilename(conn_n int, local_info, remote_info string) string {
  return fmt.Sprintf("log-%s-%04d-%s-%s.log", format_time(time.Now()), conn_n, local_info, remote_info)
}

func NewConnectionLogger(conn_n int, local_info, remote_info string) *Logger {
  return NewLogger(ConnectionLoggerFilename(conn_n, local_info, remote_info))
}

func BinaryLoggerFilename(conn_n int, peer string) string {
  return fmt.Sprintf("log-binary-%s-%04d-%s.log", format_time(time.Now()), conn_n, peer)
}

func NewBinaryLogger(conn_n int, peer string) *Logger {
  return NewLogger(BinaryLoggerFilename(conn_n, peer))
}

func (logger Logger) Log(format string, v ...interface{}) {
  logger.LogBinary([]byte(fmt.Sprintf(format + "\n", v...)))
}

func(logger Logger) LogBinary(bytes []byte) {
  logger.data <- bytes
}

func (logger Logger) LoggerLoop() {
  f, err := os.Create(logger.filename)
  if err != nil {
    die("Unable to create log file, %s, %v", logger.filename, err)
  }

  defer f.Close()

  for {
    b := <- logger.data
    if len(b) == 0 {
      break
    }

    f.Write(b)
    f.Sync()
  }
}

func (logger Logger) Close() {
  logger.data <- []byte{}
}

func format_time(t time.Time) string {
  return t.Format("2006.01.02-15.04.05")
}

func printable_addr(a net.Addr) string {
  return strings.Replace(a.String(), ":", "-", -1)
}

type Channel struct {
  from,   to            net.Conn
  logger, binary_logger *Logger
  ack                   chan bool
}

func NewChannel(conn_n int, peer string, from, to net.Conn, logger *Logger, ack chan bool) *Channel {
  channel := &Channel{ from: from, to: to, logger: logger, binary_logger: NewBinaryLogger(conn_n, peer), ack: ack }
  go channel.PassThrough()
  return channel
}

func (channel Channel) Log(format string, v ...interface{}) {
  channel.logger.Log(format, v...)
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

func process_connection(local net.Conn, conn_n int, target string) {
  remote, err := net.Dial("tcp", target)
  if err != nil {
    die("Unable to connect to %s, %v", target, err)
  }

  local_info := printable_addr(remote.LocalAddr())
  remote_info := printable_addr(remote.RemoteAddr())

  started := time.Now()

  logger := NewConnectionLogger(conn_n, local_info, remote_info)
  ack := make(chan bool)

  logger.Log("Connected to %s at %s\n", target, format_time(started))

  NewChannel(conn_n, local_info, local, remote, logger, ack)
  NewChannel(conn_n, remote_info, remote, local, logger, ack)

  // Wait for acks from *both* the pass through channels.
  <-ack
  <-ack

  finished := time.Now()
  duration := finished.Sub(started)

  logger.Log("Disconnected from %s at %s, duration %s\n", target, format_time(finished), duration.String())

  logger.Close()
}

func main() {
  runtime.GOMAXPROCS(runtime.NumCPU())
  flag.Parse()
  if flag.NFlag() != 3 {
    warn("Usage: tcpip-proxy -host target_host -port target_port -listen_port local_port")
    flag.PrintDefaults()
    os.Exit(1)
  }

  target := net.JoinHostPort(*host, *port)
  fmt.Printf("Start listening on port %s and forwarding data to %s\n", *listen_port, target)

  ln, err := net.Listen("tcp", ":" + *listen_port)
  if err != nil {
    die("Unable to start listener %v", err)
  }

  conn_n := 1
  for {
    conn, err := ln.Accept()
    if err != nil {
      warn("Accept failed: %v", err)
      continue
    }

    go process_connection(conn, conn_n, target)

    conn_n += 1
  }
}

