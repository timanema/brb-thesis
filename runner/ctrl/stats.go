package ctrl

import "time"

type Stats struct {
	Latency                            time.Duration
	MsgCount                           int
	RelayCnt, MinRelayCnt, MaxRelayCnt int
	MeanRelayCount                     float64
	BDMessagedMerged                   int
	BytesTransmitted                   int
}
