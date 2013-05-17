package tcpip_proxy

import (
  "fmt"
  "net"
  "os"
  "strings"
  "time"
)

type Log struct {
  filename string
  data     chan []byte
}

func NewConnectionLog(connectionNumber int, localInfo, remoteAddr net.Addr) *Log {
  return newLog(connectionLogFilename(connectionNumber, localInfo, remoteAddr))
}

func NewBinaryLog(connectionNumber int, peerAddr net.Addr) *Log {
  return newLog(binaryLogFilename(connectionNumber, peerAddr))
}

func (logger Log) LogLoop() {
  f, err := os.Create(logger.filename)
  if err != nil {
    panic(fmt.Sprintf("Unable to create log file, %s, %v", logger.filename, err))
  }

  defer f.Close()

  for {
    b := <-logger.data
    if len(b) == 0 {
      break
    }

    f.Write(b)
    f.Sync()
  }
}

func (logger Log) Log(format string, v ...interface{}) {
  logger.LogBinary([]byte(fmt.Sprintf("["+timestamp()+"] "+format+"\n", v...)))
}

func (logger Log) LogBinary(bytes []byte) {
  logger.data <- bytes
}

func (logger Log) Close() {
  logger.data <- []byte{}
}

func newLog(filename string) *Log {
  return &Log{
    data:     make(chan []byte),
    filename: filename,
  }
}

func connectionLogFilename(connectionNumber int, localAddr, remoteAddr net.Addr) string {
  return fmt.Sprintf("log-%s-%04d-%s-%s.log", timestamp(), connectionNumber, printableAddr(localAddr), printableAddr(remoteAddr))
}

func binaryLogFilename(connectionNumber int, peerAddr net.Addr) string {
  return fmt.Sprintf("log-binary-%s-%04d-%s.log", timestamp(), connectionNumber, printableAddr(peerAddr))
}

func timestamp() string {
  return formatTime(time.Now())
}

func formatTime(t time.Time) string {
  return t.Format("2006.01.02-15.04.05")
}

func printableAddr(a net.Addr) string {
  return strings.Replace(a.String(), ":", "-", -1)
}
