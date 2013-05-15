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

func NewConnectionLogger(connectionNumber int, localInfo, remoteInfo string) *Logger {
  return new(connectionLoggerFilename(connectionNumber, localInfo, remoteInfo))
}

func NewBinaryLogger(connectionNumber int, peer string) *Logger {
  return new(binaryLoggerFilename(connectionNumber, peer))
}

func (logger Logger) Log(format string, v ...interface{}) {
  logger.LogBinary([]byte(fmt.Sprintf("[" + timestamp() + "] " + format + "\n", v...)))
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

func connectionLoggerFilename(connectionNumber int, localInfo, remoteInfo string) string {
  return fmt.Sprintf("log-%s-%04d-%s-%s.log", timestamp(), connectionNumber, localInfo, remoteInfo)
}

func binaryLoggerFilename(connectionNumber int, peer string) string {
  return fmt.Sprintf("log-binary-%s-%04d-%s.log", timestamp(), connectionNumber, peer)
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

func timestamp() string {
  return formatTime(time.Now())
}

func formatTime(t time.Time) string {
  return t.Format("2006.01.02-15.04.05")
}
