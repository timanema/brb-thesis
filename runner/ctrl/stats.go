package ctrl

import "time"

type Stats struct {
	Latency  time.Duration
	MsgCount int
}
