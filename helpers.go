package tcpip_proxy

import (
  "time"
)

func timestamp() string {
  return formatTime(time.Now())
}

func formatTime(t time.Time) string {
  return t.Format("2006.01.02-15.04.05")
}
