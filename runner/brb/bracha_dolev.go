package brb

import (
	"fmt"
)

type brachaWrapper struct {
	messageType uint8
	msg         interface{}
}

type BrachaDolev struct {
	b *Bracha
	d *DolevImproved

	n   Network
	app Application
	cfg Config

	brachaBroadcast map[int]struct{}
}

var _ Protocol = (*BrachaDolev)(nil)
var _ Network = (*BrachaDolev)(nil)
var _ Application = (*BrachaDolev)(nil)

func (bd *BrachaDolev) Init(n Network, app Application, cfg Config) {
	bd.n = n
	bd.app = app
	bd.cfg = cfg
	bd.brachaBroadcast = make(map[int]struct{})

	sil := cfg.Silent
	if !cfg.Silent && cfg.Byz {
		fmt.Printf("process %v is a Bracha-Dolev Byzantine node\n", cfg.Id)
	}

	cfg.Silent = true

	// Create bracha instance with BD as the network
	if bd.b == nil {
		bd.b = &Bracha{}
	}
	bCfg := cfg
	nodes := cfg.Graph.Nodes()
	bCfg.Neighbours = make([]uint64, 0, nodes.Len())
	bCfg.Graph = nil

	for nodes.Next() {
		i := uint64(nodes.Node().ID())
		if i != cfg.Id {
			bCfg.Neighbours = append(bCfg.Neighbours, i)
		}
	}

	bd.b.Init(bd, app, bCfg)

	// Create dolev (improved) instance with BD as the application
	if bd.d == nil {
		bd.d = &DolevImproved{}
	}
	bd.d.Init(n, bd, cfg)

	cfg.Silent = sil
}

func (bd *BrachaDolev) Send(messageType uint8, _ uint64, uid uint32, data interface{}, bc BroadcastInfo) {
	if _, ok := bd.brachaBroadcast[bc.Id]; !ok {
		// Bracha is sending a message through Dolev
		bd.d.Broadcast(uid, brachaWrapper{
			messageType: messageType,
			msg:         data,
		}, BroadcastInfo{})
		//fmt.Printf("proc %v is broadcasting %v (%v for %v) through dolev with type %v\n", bd.cfg.Id, data, reflect.TypeOf(data).Name(), src, messageType)

		// A message is broadcast only once to all
		bd.brachaBroadcast[bc.Id] = struct{}{}
	}
}

func (bd *BrachaDolev) Deliver(uid uint32, payload interface{}, src uint64) {
	if src == bd.cfg.Id {
		return
	}

	// Dolev is delivering a message, so send it to Bracha
	m := payload.(brachaWrapper)
	bd.b.Receive(m.messageType, src, uid, m.msg)
	//fmt.Printf("proc %v has Dolev delivered %v (%v from %v) through dolev with type %v\n", bd.cfg.Id, m.msg, reflect.TypeOf(m.msg).Name(), src, m.messageType)
}

func (bd *BrachaDolev) Receive(_ uint8, src uint64, uid uint32, data interface{}) {
	// Network is delivering a messages, pass to Dolev
	bd.d.Receive(0, src, uid, data)
}

func (bd *BrachaDolev) Broadcast(uid uint32, payload interface{}, _ BroadcastInfo) {
	// Application is requesting a broadcast, pass to Bracha
	bd.b.Broadcast(uid, payload, BroadcastInfo{})
}

func (bd *BrachaDolev) Category() ProtocolCategory {
	return BrachaDolevCat
}

func (bd *BrachaDolev) TriggerStat(uid uint32, n NetworkStat) {
	bd.n.TriggerStat(uid, n)
}
