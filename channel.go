package main

import (
  "encoding/hex"
  "net"
)

type Channel struct {
  from, to             net.Conn
  connectionLogger     *Logger
  binaryLogger         *Logger
  ack                  chan bool
  buffer               []byte
  offset, packetNumber int
}

func NewChannel(from, to net.Conn, peerAddr net.Addr, connectionNumber int, connectionLogger *Logger, ack chan bool) *Channel {
  binaryLogger := NewBinaryLogger(connectionNumber, peerAddr)

  return &Channel{
    from:             from,
    to:               to,
    connectionLogger: connectionLogger,
    binaryLogger:     binaryLogger,
    ack:              ack,
    buffer:           make([]byte, 10240),
  }
}

func (channel Channel) PassThrough() {
  go channel.binaryLogger.LoggerLoop()

  for {
    err := channel.processPacket()
    if err != nil {
      break
    }
  }

  channel.disconnect()
}

func (channel Channel) log(format string, v ...interface{}) {
  channel.connectionLogger.Log(format, v...)
}

func (channel Channel) logHex(bytes []byte) {
  channel.log(hex.Dump(bytes))
}

func (channel Channel) logBinary(bytes []byte) {
  channel.binaryLogger.LogBinary(bytes)
}

func (channel Channel) read(buffer []byte) (n int, err error) {
  return channel.from.Read(buffer)
}

func (channel Channel) write(buffer []byte) (n int, err error) {
  return channel.to.Write(buffer)
}

func (channel Channel) disconnect() {
  channel.log("Disconnected from %v", channel.fromAddr())
  channel.from.Close()
  channel.to.Close()
  channel.binaryLogger.Close()
  channel.ack <- true
}

func (channel Channel) fromAddr() (addr net.Addr) {
  return channel.from.LocalAddr()
}

func (channel Channel) toAddr() (addr net.Addr) {
  return channel.to.LocalAddr()
}

func (channel Channel) processPacket() error {
  n, err := channel.read(channel.buffer)
  if err == nil && n > 0 {
    channel.processSuccessfulPacket(n)
  }
  return err
}

func (channel Channel) processSuccessfulPacket(bytesRead int) {
  channel.log("Received (#%d, %08X) %d bytes from %v", channel.packetNumber, channel.offset, bytesRead, channel.fromAddr())
  channel.logAndWriteData(channel.buffer[:bytesRead])
  channel.log("Sent (#%d) to %v\n", channel.packetNumber, channel.toAddr())

  channel.offset += bytesRead
  channel.packetNumber += 1
}

func (channel Channel) logAndWriteData(data []byte) {
  channel.logHex(data)
  channel.logBinary(data)
  channel.write(data)
}
