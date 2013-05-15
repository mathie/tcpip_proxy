package main

import (
  "flag"
  "fmt"
  "net"
  "os"
  "runtime"
  "connection"
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

func (proxy Proxy) ProcessConnection(incomingConnection net.Conn) {
  connection.NewConnection(incomingConnection, proxy.connectionNumber, proxy.target)

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
