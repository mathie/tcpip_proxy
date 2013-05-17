package tcpip_proxy

import (
  "fmt"
  "net"
  "strings"
  "time"
)

const (
  LocalToRemote = iota
  RemoteToLocal
)

type Connection struct {
  local, remote    net.Conn
  connectionNumber int
  target           string
  logger           Logger
  ack              chan bool
}

func NewConnection(local net.Conn, connectionNumber int, target string) *Connection {
  remote, err := net.Dial("tcp", target)
  if err != nil {
    panic(fmt.Sprintf("Unable to connect to %s, %v", target, err))
  }

  return &Connection{
    local:            local,
    remote:           remote,
    connectionNumber: connectionNumber,
    target:           target,
    ack:              make(chan bool),
  }
}

func (connection Connection) Process() {
  connectionLog := NewConnectionLog(connectionLogFilename(connection.connectionNumber, connection.localAddr(), connection.remoteAddr()))
  go connectionLog.LogLoop()
  defer connectionLog.Close()

  started := time.Now()

  connectionLog.Log("Connected to %s.\n", connection.target)

  localToRemoteChannel := connection.newChannel(LocalToRemote, connectionLog)
  remoteToLocalChannel := connection.newChannel(RemoteToLocal, connectionLog)

  go localToRemoteChannel.PassThrough()
  go remoteToLocalChannel.PassThrough()

  // Wait for acks from *both* the pass through channels.
  <-connection.ack
  <-connection.ack

  finished := time.Now()
  duration := finished.Sub(started)

  connectionLog.Log("Disconnected from %s, duration %s.\n", connection.target, duration.String())
}

func (connection Connection) localAddr() net.Addr {
  return connection.remote.LocalAddr()
}

func (connection Connection) remoteAddr() net.Addr {
  return connection.remote.RemoteAddr()
}

func (connection Connection) channelAddr(direction int) net.Addr {
  switch direction {
  case LocalToRemote:
    return connection.localAddr()
  case RemoteToLocal:
    return connection.remoteAddr()
  }

  panic("Unreachable.")
}

func (connection Connection) from(direction int) net.Conn {
  switch direction {
  case LocalToRemote:
    return connection.local
  case RemoteToLocal:
    return connection.remote
  }

  panic("Unreachable.")
}

func (connection Connection) to(direction int) net.Conn {
  switch direction {
  case LocalToRemote:
    return connection.remote
  case RemoteToLocal:
    return connection.local
  }

  panic("Unreachable.")
}

func (connection Connection) newChannel(direction int, connectionLog Logger) *Channel {
  binaryLog := NewBinaryLog(binaryLogFilename(connection.connectionNumber, connection.channelAddr(direction)))

  return NewChannel(connection.from(direction), connection.to(direction), binaryLog, connectionLog, connection.ack)
}

func connectionLogFilename(connectionNumber int, localAddr, remoteAddr net.Addr) string {
  return fmt.Sprintf("log-%s-%04d-%s-%s.log", timestamp(), connectionNumber, printableAddr(localAddr), printableAddr(remoteAddr))
}

func binaryLogFilename(connectionNumber int, peerAddr net.Addr) string {
  return fmt.Sprintf("log-binary-%s-%04d-%s.log", timestamp(), connectionNumber, printableAddr(peerAddr))
}

func printableAddr(a net.Addr) string {
  return strings.Replace(a.String(), ":", "-", -1)
}
