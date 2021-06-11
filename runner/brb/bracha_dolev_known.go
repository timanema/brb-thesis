package brb

import (
	"fmt"
	"reflect"
	"rp-runner/brb/algo"
	"rp-runner/graphs"
)

type BrachaDolevConfig struct {
	Included algo.BrachaInclusionTable
}

type BrachaDolevMessage struct {
	Src   uint64
	Id    uint32
	Type  uint8
	Paths []algo.DolevPath
}

func (b BrachaDolevMessage) SizeOf() uintptr {
	return reflect.TypeOf(b.Src).Size() + reflect.TypeOf(b.Id).Size() + reflect.TypeOf(b.Type).Size() + algo.SizeOfMultiplePaths(b.Paths)
}

// BrachaDolevKnown can be used to compare naive routing to improved routing
type BrachaDolevKnown struct {
	wr *brachaDolevKnownWrapper
}

var _ Protocol = (*BrachaDolevKnown)(nil)
var _ Network = (*BrachaDolevKnown)(nil)
var _ Application = (*BrachaDolevKnown)(nil)

func (bd *BrachaDolevKnown) Init(n Network, app Application, cfg Config) {
	if bd.wr == nil {
		bd.wr = &brachaDolevKnownWrapper{bracha: &BrachaImproved{}, dolev: &DolevKnown{}}
	}

	bd.wr.Init(n, app, cfg)
}

func (bd *BrachaDolevKnown) Send(messageType uint8, dest uint64, uid uint32, data Size, bc BroadcastInfo) {
	bd.wr.Send(messageType, dest, uid, data, bc)
}

func (bd *BrachaDolevKnown) Deliver(uid uint32, payload Size, src uint64) {
	bd.wr.Deliver(uid, payload, src)
}

func (bd *BrachaDolevKnown) Receive(messageType uint8, src uint64, uid uint32, data Size) {
	bd.wr.Receive(messageType, src, uid, data)
}

func (bd *BrachaDolevKnown) Broadcast(uid uint32, payload Size, bc BroadcastInfo) {
	bd.wr.Broadcast(uid, payload, bc)
}

func (bd *BrachaDolevKnown) Category() ProtocolCategory {
	return bd.wr.Category()
}

func (bd *BrachaDolevKnown) TriggerStat(uid uint32, n NetworkStat) {
	bd.wr.TriggerStat(uid, n)
}

// BrachaDolevKnownImproved uses improved routing
type BrachaDolevKnownImproved struct {
	wr *brachaDolevKnownWrapper
}

var _ Protocol = (*BrachaDolevKnownImproved)(nil)
var _ Network = (*BrachaDolevKnownImproved)(nil)
var _ Application = (*BrachaDolevKnownImproved)(nil)

func (bd *BrachaDolevKnownImproved) Init(n Network, app Application, cfg Config) {
	if bd.wr == nil {
		bd.wr = &brachaDolevKnownWrapper{bracha: &BrachaImproved{}, dolev: &DolevKnownImprovedPM{}}
	}

	bd.wr.Init(n, app, cfg)
}

func (bd *BrachaDolevKnownImproved) Send(messageType uint8, dest uint64, uid uint32, data Size, bc BroadcastInfo) {
	bd.wr.Send(messageType, dest, uid, data, bc)
}

func (bd *BrachaDolevKnownImproved) Deliver(uid uint32, payload Size, src uint64) {
	bd.wr.Deliver(uid, payload, src)
}

func (bd *BrachaDolevKnownImproved) Receive(messageType uint8, src uint64, uid uint32, data Size) {
	bd.wr.Receive(messageType, src, uid, data)
}

func (bd *BrachaDolevKnownImproved) Broadcast(uid uint32, payload Size, bc BroadcastInfo) {
	bd.wr.Broadcast(uid, payload, bc)
}

func (bd *BrachaDolevKnownImproved) Category() ProtocolCategory {
	return bd.wr.Category()
}

func (bd *BrachaDolevKnownImproved) TriggerStat(uid uint32, n NetworkStat) {
	bd.wr.TriggerStat(uid, n)
}

type brachaDolevKnownWrapper struct {
	bracha Protocol
	dolev  Protocol

	n   Network
	app Application
	cfg Config

	brachaBroadcast map[int]struct{}
}

func (bd *brachaDolevKnownWrapper) Init(n Network, app Application, cfg Config) {
	bd.n = n
	bd.app = app
	bd.cfg = cfg
	bd.brachaBroadcast = make(map[int]struct{})

	sil := cfg.Silent
	if !cfg.Silent && cfg.Byz {
		fmt.Printf("process %v is a Bracha-Dolev Byzantine node\n", cfg.Id)
	}

	cfg.Silent = true

	bCfg := cfg
	nodes := cfg.Graph.Nodes()
	bCfg.Neighbours = make([]uint64, 0, nodes.Len())

	nids, _ := graphs.Nodes(cfg.Graph)

	bCfg.AdditionalConfig = BrachaDolevConfig{
		Included: algo.FindBrachaDolevInclusionTable(cfg.Graph, nids, cfg.N, cfg.F),
	}

	for nodes.Next() {
		i := uint64(nodes.Node().ID())
		if i != cfg.Id {
			bCfg.Neighbours = append(bCfg.Neighbours, i)
		}
	}

	bd.bracha.Init(bd, app, bCfg)

	// Create dolev (improved) instance with BD as the application
	if bd.dolev == nil {
		bd.dolev = &DolevKnownImprovedPM{}
	}
	cfg.AdditionalConfig = BrachaDolevConfig{}
	bd.dolev.Init(n, bd, cfg)

	cfg.Silent = sil
}

func (bd *brachaDolevKnownWrapper) Send(messageType uint8, dest uint64, uid uint32, data Size, bc BroadcastInfo) {
	if _, ok := bd.brachaBroadcast[bc.Id]; !ok {
		// A message is broadcast only once to all
		bd.brachaBroadcast[bc.Id] = struct{}{}

		//if messageType == BrachaEcho {
		//	fmt.Printf("proc %v is sending echo\n", bd.cfg.Id)
		//} else if messageType == BrachaReady {
		//	fmt.Printf("proc %v is sending ready\n", bd.cfg.Id)
		//}

		partial := messageType == BrachaSend || messageType == BrachaEcho
		bcType := BrachaEveryone

		if partial {
			bcType = BrachaPartial
		}

		// Bracha is sending a message through Dolev
		bd.dolev.Broadcast(uid, brachaWrapper{
			messageType: messageType,
			msg:         data,
		}, BroadcastInfo{Type: bcType})
	}
}

func (bd *brachaDolevKnownWrapper) Deliver(uid uint32, payload Size, src uint64) {
	// Dolev is delivering a message, so send it to Bracha
	m := payload.(brachaWrapper)

	if src == bd.cfg.Id {
		return
	}

	//if m.messageType == BrachaSend {
	//	fmt.Printf("proc %v got initial send\n", bd.cfg.Id)
	//} else if m.messageType == BrachaEcho {
	//	fmt.Printf("proc %v got echo from %v\n", bd.cfg.Id, src)
	//} else if m.messageType == BrachaReady {
	//	fmt.Printf("proc %v got ready from %v\n", bd.cfg.Id, src)
	//}

	bd.bracha.Receive(m.messageType, src, uid, m.msg)
}

func (bd *brachaDolevKnownWrapper) Receive(_ uint8, src uint64, uid uint32, data Size) {
	// Network is delivering a messages, pass to Dolev
	bd.dolev.Receive(0, src, uid, data)
}

func (bd *brachaDolevKnownWrapper) Broadcast(uid uint32, payload Size, _ BroadcastInfo) {
	// Application is requesting a broadcast, pass to Bracha
	bd.bracha.Broadcast(uid, payload, BroadcastInfo{})
}

func (bd *brachaDolevKnownWrapper) Category() ProtocolCategory {
	return BrachaDolevCat
}

func (bd *brachaDolevKnownWrapper) TriggerStat(uid uint32, n NetworkStat) {
	bd.n.TriggerStat(uid, n)
}
