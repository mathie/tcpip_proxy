package tcpip_proxy

import (
  "encoding/hex"
  "io"
  "net"
)

type Addressable interface {
  LocalAddr() net.Addr
}

type ReadAddressableCloser interface {
  io.ReadCloser
  Addressable
}

type WriteAddressableCloser interface {
  io.WriteCloser
  Addressable
}

// Represents a unidirectional channel between two TCP sockets. A channel only
// deals with communication in one direction, so a Connection (which
// encapsulates a Channel in each direction) is generally sought. The channel
// reads data from one side, logs it to a binary logger and writes it to the
// other side.
type Channel struct {
  from                 ReadAddressableCloser
  to                   WriteAddressableCloser
  connectionLog        Logger
  binaryLog            BinaryLogger
  ack                  chan bool
  buffer               []byte
  offset, packetNumber int
}

func NewChannel(from ReadAddressableCloser, to WriteAddressableCloser, binaryLog BinaryLogger, connectionLog Logger, ack chan bool) *Channel {
  return &Channel{
    from:          from,
    to:            to,
    connectionLog: connectionLog,
    binaryLog:     binaryLog,
    ack:           ack,
    buffer:        make([]byte, 10240),
  }
}

func (channel Channel) PassThrough() {
  go channel.binaryLog.LogLoop()

  for {
    err := channel.processPacket()
    if err != nil {
      break
    }
  }

  channel.disconnect()
}

func (channel Channel) log(format string, v ...interface{}) {
  channel.connectionLog.Log(format, v...)
}

func (channel Channel) logHex(bytes []byte) {
  channel.log(hex.Dump(bytes))
}

func (channel Channel) logBinary(bytes []byte) {
  channel.binaryLog.LogBinary(bytes)
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
  channel.binaryLog.Close()
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
