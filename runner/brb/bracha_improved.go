package brb

import (
	"crypto/sha256"
	"fmt"
	"math"
)

// Improved version for RP Tim Anema
type BrachaImproved struct {
	n   Network
	app Application
	cfg Config

	delivered map[uint32]struct{}

	echo  map[brachaIdentifier]map[uint64]struct{}
	ready map[brachaIdentifier]map[uint64]struct{}

	echoSent  map[brachaIdentifier]struct{}
	readySent map[brachaIdentifier]struct{}
}

func (b *BrachaImproved) Init(n Network, app Application, cfg Config) {
	b.n = n
	b.app = app
	b.cfg = cfg
	b.delivered = make(map[uint32]struct{})
	b.echo = make(map[brachaIdentifier]map[uint64]struct{})
	b.ready = make(map[brachaIdentifier]map[uint64]struct{})
	b.echoSent = make(map[brachaIdentifier]struct{})
	b.readySent = make(map[brachaIdentifier]struct{})

	if cfg.Byz {
		fmt.Printf("process %v is a Bracha Byzantine node\n", cfg.Id)
	}
}

func (b *BrachaImproved) send(messageType uint8, uid uint32, id brachaIdentifier, data interface{}) {
	// Only send one echo per message
	if _, ok := b.echoSent[id]; ok && messageType == BrachaEcho {
		return
	}

	// Only send one ready per message
	if _, ok := b.readySent[id]; ok && messageType == BrachaReady {
		return
	}

	for _, n := range b.cfg.Neighbours {
		b.n.Send(messageType, n, uid, data)
	}
}

func (b *BrachaImproved) hasDelivered(uid uint32) bool {
	_, ok := b.delivered[uid]
	return ok
}

func (b *BrachaImproved) Receive(messageType uint8, src uint64, uid uint32, data interface{}) {
	if b.cfg.Byz {
		// TODO: better byzantine behaviour?
		return
	}

	m := data.(BrachaMessage)

	id := brachaIdentifier{
		Id:   uid,
		Hash: sha256.Sum256(m.Payload),
	}

	_, echoMade := b.echo[id]
	_, readyMade := b.ready[id]
	if !echoMade || !readyMade {
		b.echo[id] = make(map[uint64]struct{})
		b.ready[id] = make(map[uint64]struct{})
	}

	del := b.hasDelivered(uid)
	switch messageType {
	case BrachaSend:
		b.send(BrachaEcho, uid, id, data)
		b.echoSent[id] = struct{}{}
	case BrachaEcho:
		if !del {
			b.echo[id][src] = struct{}{}
		}
	case BrachaReady:
		if !del {
			b.ready[id][src] = struct{}{}
		}
	}

	// Send ready if enough ((n + f + 1) / 2) echos, or if enough readys
	if len(b.echo[id]) > int(math.Ceil((float64(b.cfg.N)+float64(b.cfg.F)+1)/2)) || len(b.ready[id]) >= b.cfg.F+1 {
		b.send(BrachaReady, uid, id, data)
		b.readySent[id] = struct{}{}
	}

	// Deliver if enough readys
	if !b.hasDelivered(uid) && len(b.ready[id]) >= b.cfg.F*2+1 {
		b.delivered[uid] = struct{}{}
		b.app.Deliver(uid, m.Payload)

		// Memory cleanup
		delete(b.echo, id)
		delete(b.ready, id)
		delete(b.echoSent, id)
		delete(b.readySent, id)
	}
}

func (b *BrachaImproved) Send(uid uint32, payload []byte) {
	if _, ok := b.delivered[uid]; !ok {
		id := brachaIdentifier{
			Id:   uid,
			Hash: sha256.Sum256(payload),
		}

		b.echo[id] = map[uint64]struct{}{
			b.cfg.Id: {},
		}
		b.ready[id] = map[uint64]struct{}{
			b.cfg.Id: {},
		}

		m := BrachaMessage{
			Src:     b.cfg.Id,
			Payload: payload,
		}

		b.send(BrachaSend, uid, id, m)
	}
}
