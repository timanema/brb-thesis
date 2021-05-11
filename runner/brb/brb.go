package brb

import (
	"gonum.org/v1/gonum/graph/simple"
)

// Used as abstraction for BRB protocols
// uid is used for tracking the message throughout the network (for statistics)

type Application interface {
	Deliver(uid uint32, payload []byte)
}

type Network interface {
	Send(messageType uint8, dest uint64, uid uint32, data []byte)
}

type Config struct {
	Byz           bool
	N, F          int
	Id            uint64
	Neighbours    []uint64
	Graph         *simple.WeightedUndirectedGraph
	KnownTopology bool
}

type Protocol interface {
	// Can be used to do some initial work (processes will wait for this to
	// be completed before announcing that they're ready)
	Init(n Network, app Application, cfg Config)

	Receive(messageType uint8, src uint64, uid uint32, data []byte)

	Send(uid uint32, payload []byte)
}
