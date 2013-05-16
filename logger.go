package main

import (
  "fmt"
  "net"
  "os"
  "strings"
  "time"
)

type Logger struct {
  filename string
  data     chan []byte
}

func NewConnectionLogger(connectionNumber int, localInfo, remoteAddr net.Addr) *Logger {
  return newLogger(connectionLoggerFilename(connectionNumber, localInfo, remoteAddr))
}

func NewBinaryLogger(connectionNumber int, peerAddr net.Addr) *Logger {
  return newLogger(binaryLoggerFilename(connectionNumber, peerAddr))
}

func (logger Logger) LoggerLoop() {
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

func (logger Logger) Log(format string, v ...interface{}) {
  logger.LogBinary([]byte(fmt.Sprintf("["+timestamp()+"] "+format+"\n", v...)))
}

func (logger Logger) LogBinary(bytes []byte) {
  logger.data <- bytes
}

func (logger Logger) Close() {
  logger.data <- []byte{}
}

func newLogger(filename string) *Logger {
  return &Logger{
    data:     make(chan []byte),
    filename: filename,
  }
}

func connectionLoggerFilename(connectionNumber int, localAddr, remoteAddr net.Addr) string {
  return fmt.Sprintf("log-%s-%04d-%s-%s.log", timestamp(), connectionNumber, printableAddr(localAddr), printableAddr(remoteAddr))
}

func binaryLoggerFilename(connectionNumber int, peerAddr net.Addr) string {
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
