package tcpip_proxy

import (
  "fmt"
  "net"
)

type Proxy struct {
  target           string
  localPort        string
  connectionNumber int
}

func RunProxy(targetHost, targetPort, localPort string) {
  target := net.JoinHostPort(targetHost, targetPort)

  proxy := &Proxy{
    target:           target,
    localPort:        localPort,
    connectionNumber: 1,
  }

  proxy.run()
}

func (proxy Proxy) run() {
  fmt.Printf("Start listening on port %s and forwarding data to %s\n", proxy.localPort, proxy.target)

  ln, err := net.Listen("tcp", ":"+proxy.localPort)
  if err != nil {
    panic(fmt.Sprintf("Unable to start listener %v", err))
  }

  for {
    conn, err := ln.Accept()
    if err != nil {
      panic(fmt.Sprintf("Accept failed: %v", err))
    }

    proxy.processConnection(conn)
  }
}

func (proxy Proxy) processConnection(incomingConnection net.Conn) {
  newConnection := NewConnection(incomingConnection, proxy.connectionNumber, proxy.target)
  go newConnection.Process()

  proxy.connectionNumber += 1
}
