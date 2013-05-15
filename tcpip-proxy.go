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

func connection_logger(data chan []byte, conn_n int, local_info, remote_info string) {
  log_name := fmt.Sprintf("log-%s-%04d-%s-%s.log", format_time(time.Now()), conn_n, local_info, remote_info)

  logger_loop(data, log_name)
}

func binary_logger(data chan []byte, conn_n int, peer string) {
  log_name := fmt.Sprintf("log-binary-%s-%04d-%s.log", format_time(time.Now()), conn_n, peer)

  logger_loop(data, log_name)
}

func logger_loop(data chan []byte, log_name string) {
  f, err := os.Create(log_name)
  if err != nil {
    die("Unable to create log file, %s, %v", log_name, err)
  }

  defer f.Close()

  for {
    b := <-data
    if len(b) == 0 {
      break
    }

    f.Write(b)
    f.Sync()
  }
}

func format_time(t time.Time) string {
  return t.Format("2006.01.02-15.04.05")
}

func printable_addr(a net.Addr) string {
  return strings.Replace(a.String(), ":", "-", -1)
}

func log(logger chan []byte, format string, v ...interface{}) {
  logger <- []byte(fmt.Sprintf(format + "\n", v...))
}

type Channel struct {
  from,   to            net.Conn
  logger, binary_logger chan []byte
  ack                   chan bool
}

func (channel Channel) ChanLog(format string, v ...interface{}) {
  log(channel.logger, format, v...)
}

func (channel Channel) LogHex(bytes []byte) {
  channel.logger <- []byte(hex.Dump(bytes))
}

func (channel Channel) LogBinary(bytes []byte) {
  channel.binary_logger <- bytes
}

func (channel Channel) Read(buffer []byte) (n int, err error) {
  return channel.from.Read(buffer)
}

func (channel Channel) Write(buffer []byte) (n int, err error) {
  return channel.to.Write(buffer)
}

func (channel Channel) FromAddr() (addr string) {
  return printable_addr(channel.from.LocalAddr())
}

func (channel Channel) ToAddr() (addr string) {
  return printable_addr(channel.to.LocalAddr())
}

func pass_through(c *Channel) {
  b := make([]byte, 10240)
  offset := 0
  packet_n := 0

  for {
    n, err := c.Read(b)
    if err != nil {
      break
    }

    if n <= 0 {
      continue
    }

    c.ChanLog("Received (#%d, %08X) %d bytes from %s", packet_n, offset, n, c.FromAddr())

    c.LogHex(b[:n])
    c.LogBinary(b[:n])

    c.Write(b[:n])

    c.ChanLog("Sent (#%d) to %s\n", packet_n, c.ToAddr())

    offset += n
    packet_n += 1
  }

  c.ChanLog("Disconnected from %s", c.FromAddr())
  c.from.Close()
  c.to.Close()
  c.ack <- true
}

func close_logger(logger chan []byte) {
  logger <- []byte{}
}

func process_connection(local net.Conn, conn_n int, target string) {
  remote, err := net.Dial("tcp", target)
  if err != nil {
    die("Unable to connect to %s, %v", target, err)
  }

  local_info := printable_addr(remote.LocalAddr())
  remote_info := printable_addr(remote.RemoteAddr())

  started := time.Now()

  logger := make(chan []byte)
  from_logger := make(chan []byte)
  to_logger := make(chan []byte)
  ack := make(chan bool)

  go connection_logger(logger, conn_n, local_info, remote_info)
  go binary_logger(from_logger, conn_n, local_info)
  go binary_logger(to_logger, conn_n, remote_info)

  log(logger, "Connected to %s at %s\n", target, format_time(started))

  go pass_through(&Channel{remote, local, logger, to_logger, ack})
  go pass_through(&Channel{local, remote, logger, from_logger, ack})

  // Wait for acks from *both* the pass through channels.
  <-ack
  <-ack

  finished := time.Now()
  duration := finished.Sub(started)

  log(logger, "Disconnected from %s at %s, duration %s\n", target, format_time(finished), duration.String())

  close_logger(logger)
  close_logger(from_logger)
  close_logger(to_logger)
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
    fmt.Printf("Unable to start listener %v\n", err)
    os.Exit(1)
  }

  conn_n := 1
  for {
    conn, err := ln.Accept()
    if err != nil {
      fmt.Printf("Accept failed: %v\n", err)
      continue
    }

    go process_connection(conn, conn_n, target)

    conn_n += 1
  }
}

