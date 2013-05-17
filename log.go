package tcpip_proxy

import (
  "fmt"
  "os"
)

type LogChannel interface {
  LogLoop()
  Close()
}

type Logger interface {
  LogChannel
  Log(format string, v ...interface{})
}

type BinaryLogger interface {
  LogChannel
  LogBinary(bytes []byte)
}

type Log struct {
  filename string
  data     chan []byte
}

func NewConnectionLog(filename string) Logger {
  return newLog(filename)
}

func NewBinaryLog(filename string) BinaryLogger {
  return newLog(filename)
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
