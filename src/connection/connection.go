package connection

import (
  "net"
  "fmt"
  "time"
  "logger"
  "channel"
)

const (
  LocalToRemote = iota
  RemoteToLocal
)

type Connection struct {
  local, remote net.Conn
  connectionNumber int
  target string
  logger *logger.Logger
  ack chan bool
}

func NewConnection(local net.Conn, connectionNumber int, target string) *Connection {
  remote, err := net.Dial("tcp", target)
  if err != nil {
    panic(fmt.Sprintf("Unable to connect to %s, %v", target, err))
  }

  connection := &Connection{ local: local, remote: remote, connectionNumber: connectionNumber, target: target, ack: make(chan bool) }
  go connection.process()
  return connection
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

func (connection Connection) newChannel(direction int, connectionLogger *logger.Logger) *channel.Channel {
  return channel.NewChannel(connection.from(direction), connection.to(direction), connection.channelAddr(direction), connection.connectionNumber, connectionLogger, connection.ack)
}

func (connection Connection) process() {
  connectionLogger := logger.NewConnectionLogger(connection.connectionNumber, connection.localAddr(), connection.remoteAddr())
  defer connectionLogger.Close()

  started := time.Now()

  connectionLogger.Log("Connected to %s.\n", connection.target)

  connection.newChannel(LocalToRemote, connectionLogger)
  connection.newChannel(RemoteToLocal, connectionLogger)

  // Wait for acks from *both* the pass through channels.
  <-connection.ack
  <-connection.ack

  finished := time.Now()
  duration := finished.Sub(started)

  connectionLogger.Log("Disconnected from %s, duration %s.\n", connection.target, duration.String())
}