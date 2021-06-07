package brb

import (
	"gonum.org/v1/gonum/graph/simple"
)

// Used as abstraction for BRB protocols
// uid is used for tracking the message throughout the network (for statistics)

type Size interface {
	SizeOf() uintptr
}

type Application interface {
	Deliver(uid uint32, payload Size, src uint64)
}

type BroadcastInfo struct {
	Type, Id int
}

const (
	BrachaEveryone int = iota
	BrachaPartial
)

type Network interface {
	// BroadcastInfo can be used to pass information between protocols, regular applications and networks will not use this
	Send(messageType uint8, dest uint64, uid uint32, data Size, bc BroadcastInfo)

	TriggerStat(uid uint32, n NetworkStat)
}

type NetworkStat int

const (
	StartRelay NetworkStat = iota
	BrachaDolevMerge
	DolevPayloadMerge
	DolevPathMerge
)

type OptimizationConfig struct {
	DolevFilterSubpaths, DolevSingleHopNeighbour,
	DolevCombineNextHops, DolevReusePaths,
	DolevRelayMerging, DolevPayloadMerging, DolevImplicitPath bool
	BrachaImplicitEcho, BrachaMinimalSubset       bool
	BrachaDolevPartialBroadcast, BrachaDolevMerge bool
}

type Config struct {
	Byz        bool
	N, F       int
	Id         uint64
	Neighbours []uint64
	Graph      *simple.WeightedUndirectedGraph

	Silent, Unused bool

	AdditionalConfig   interface{}
	OptimizationConfig OptimizationConfig
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

	Receive(messageType uint8, src uint64, uid uint32, data Size)

	// BroadcastInfo can be used to pass information between protocols, regular applications and networks will not use this
	Broadcast(uid uint32, payload Size, bc BroadcastInfo)

	Category() ProtocolCategory
}
