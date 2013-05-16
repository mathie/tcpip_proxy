package main

import (
  "flag"
  "fmt"
  "os"
  "runtime"
)

var (
  host *string = flag.String("host", "", "target host or address")
  port *string = flag.String("port", "0", "target port")
  listenPort *string = flag.String("listenPort", "0", "listen port")
)

func warn(format string, v ...interface{}) {
  os.Stderr.WriteString(fmt.Sprintf(format + "\n", v...))
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

  RunProxy(*host, *port, *listenPort)
}
