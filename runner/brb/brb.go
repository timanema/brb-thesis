package brb

import (
	"gonum.org/v1/gonum/graph/simple"
)

// Used as abstraction for BRB protocols
// uid is used for tracking the message throughout the network (for statistics)

type Application interface {
	Deliver(uid uint32, payload interface{}, src uint64)
}

type BroadcastInfo struct {
	Type, Id int
}

type Network interface {
	Send(messageType uint8, dest uint64, uid uint32, data interface{}, bc BroadcastInfo)

	TriggerStat(uid uint32, n NetworkStat)
}

type NetworkStat int

const (
	StartRelay NetworkStat = iota
)

type Config struct {
	Byz           bool
	N, F          int
	Id            uint64
	Neighbours    []uint64
	Graph         *simple.WeightedUndirectedGraph
	KnownTopology bool

	Silent, Unused bool
}

type ProtocolCategory int

const (
	DolevCat ProtocolCategory = iota
	BrachaCat
	BrachaDolevCat
)

type Protocol interface {
	// Can be used to do some initial work (processes will wait for this to
	// be completed before announcing that they're ready)
	Init(n Network, app Application, cfg Config)

	Receive(messageType uint8, src uint64, uid uint32, data interface{})

	Broadcast(uid uint32, payload interface{})

	Category() ProtocolCategory
}
