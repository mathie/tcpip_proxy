package logger

import (
  "fmt"
  "os"
  "time"
)

type Logger struct {
  filename string
  data chan []byte
}

func NewConnectionLogger(conn_n int, local_info, remote_info string) *Logger {
  return new(connectionLoggerFilename(conn_n, local_info, remote_info))
}

func NewBinaryLogger(conn_n int, peer string) *Logger {
  return new(binaryLoggerFilename(conn_n, peer))
}

func (logger Logger) Log(format string, v ...interface{}) {
  logger.LogBinary([]byte(fmt.Sprintf(format + "\n", v...)))
}

func(logger Logger) LogBinary(bytes []byte) {
  logger.data <- bytes
}

func (logger Logger) Close() {
  logger.data <- []byte{}
}

func new(filename string) *Logger {
  logger := &Logger { data: make(chan []byte), filename: filename }
  go logger.loggerLoop()
  return logger
}

func connectionLoggerFilename(conn_n int, local_info, remote_info string) string {
  return fmt.Sprintf("log-%s-%04d-%s-%s.log", formatTime(time.Now()), conn_n, local_info, remote_info)
}

func binaryLoggerFilename(conn_n int, peer string) string {
  return fmt.Sprintf("log-binary-%s-%04d-%s.log", formatTime(time.Now()), conn_n, peer)
}

func (logger Logger) loggerLoop() {
  f, err := os.Create(logger.filename)
  if err != nil {
    panic(fmt.Sprintf("Unable to create log file, %s, %v", logger.filename, err))
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

func formatTime(t time.Time) string {
  return t.Format("2006.01.02-15.04.05")
}
