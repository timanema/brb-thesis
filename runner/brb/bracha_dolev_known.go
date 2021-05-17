package brb

import (
	"fmt"
)

type BrachaDolevKnown struct {
	b *BrachaImproved
	d *DolevKnownImproved

	n   Network
	app Application
	cfg Config
}

var _ Protocol = (*BrachaDolevKnown)(nil)
var _ Network = (*BrachaDolevKnown)(nil)
var _ Application = (*BrachaDolevKnown)(nil)

func (bd *BrachaDolevKnown) Init(n Network, app Application, cfg Config) {
	bd.n = n
	bd.app = app
	bd.cfg = cfg

	sil := cfg.Silent
	if !cfg.Silent && cfg.Byz {
		fmt.Printf("process %v is a Bracha-Dolev Byzantine node\n", cfg.Id)
	}

	cfg.Silent = true

	// Create bracha instance with BD as the network
	if bd.b == nil {
		bd.b = &BrachaImproved{}
	}
	bd.b.Init(bd, app, cfg)

	// Create dolev (improved) instance with BD as the application
	if bd.d == nil {
		bd.d = &DolevKnownImproved{}
	}
	bd.d.Init(n, bd, cfg)

	cfg.Silent = sil
}

func (bd *BrachaDolevKnown) Send(messageType uint8, src uint64, uid uint32, data interface{}) {
	// Bracha is sending a message through Dolev
	bd.d.Broadcast(uid, brachaWrapper{
		messageType: messageType,
		msg:         data,
	})
}

func (bd *BrachaDolevKnown) Deliver(uid uint32, payload interface{}, src uint64) {
	// Dolev is delivering a message, so send it to Bracha
	m := payload.(brachaWrapper)
	bd.b.Receive(m.messageType, src, uid, m.msg)
}

func (bd *BrachaDolevKnown) Receive(_ uint8, src uint64, uid uint32, data interface{}) {
	// Network is delivering a messages, pass to Dolev
	bd.d.Receive(0, src, uid, data)
}

func (bd *BrachaDolevKnown) Broadcast(uid uint32, payload interface{}) {
	// Application is requesting a broadcast, pass to Bracha
	bd.b.Broadcast(uid, payload)
}
