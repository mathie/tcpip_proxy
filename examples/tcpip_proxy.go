// TCP/IP Proxy implements a simple TCP proxy which listens on a local port and
// proxies to a remote host/port while logging all the communication out to a
// file. It produces a log file with a hex dump of the packets logged, and a
// binary log of the streams in each direction for each connection.
//
// Usage:
//        tcpip_proxy -host targetHost -port targetPort -listenPort listenPort
//
// The flags are:
//
//        -host=targetHost
//                the remote host to connect to when a local connection is
//                received.
//
//        -port=targetPort
//                the TCP port on the remote host to connect to when a local
//                connection is received.
//
//        -listenPort=listenPort
//                The local TCP port to listen on for connections. Proxied
//                services should connect to this local port in order to log
//                traffic to the remote host.
//
// Example:
//
//        tcpip_proxy -host www.google.com -port 80 -listenPort 8080
//
// which will create a proxy from localhost port 8080 to www.google.com:80. We
// can exercise this with:
//
//        curl -v http://localhost:8080/
//
// and see that the results are logged.
package main

import (
  "flag"
  "fmt"
  "os"
  "runtime"
  "github.com/mathie/tcpip_proxy"
)

// Command line flags.
var (
  host       *string = flag.String("host", "", "target host or address")
  port       *string = flag.String("port", "0", "target port")
  listenPort *string = flag.String("listenPort", "0", "listen port")
)

func warn(format string, v ...interface{}) {
  os.Stderr.WriteString(fmt.Sprintf(format+"\n", v...))
}

func parseArgs() {
  flag.Parse()
  if flag.NFlag() != 3 {
    warn("Usage: tcpip-proxy -host targetHost -port targetPort -listenPort localPort")
    flag.PrintDefaults()
    os.Exit(1)
  }
}

func main() {
  runtime.GOMAXPROCS(runtime.NumCPU())

  parseArgs()

  tcpip_proxy.RunProxy(*host, *port, *listenPort)
}
