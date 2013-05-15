package channel

import (
  "net"
  "encoding/hex"
  "logger"
)

type Channel struct {
  from, to            net.Conn
  connectionLogger, binaryLogger *logger.Logger
  ack                   chan bool
}

func NewChannel(from, to net.Conn, peerAddr net.Addr, connectionNumber int, connectionLogger *logger.Logger, ack chan bool) *Channel {
  binaryLogger := logger.NewBinaryLogger(connectionNumber, peerAddr)
  channel := &Channel{ from: from, to: to, connectionLogger: connectionLogger, binaryLogger: binaryLogger, ack: ack }

  go channel.passThrough()
  return channel
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

func (channel Channel) passThrough() {
  b := make([]byte, 10240)
  offset := 0
  packetNumber := 0

  for {
    n, err := channel.read(b)
    if err != nil {
      break
    }

    if n <= 0 {
      continue
    }

    channel.log("Received (#%d, %08X) %d bytes from %v", packetNumber, offset, n, channel.fromAddr())

    channel.logHex(b[:n])
    channel.logBinary(b[:n])

    channel.write(b[:n])

    channel.log("Sent (#%d) to %v\n", packetNumber, channel.toAddr())

    offset += n
    packetNumber += 1
  }

  channel.disconnect()
}
