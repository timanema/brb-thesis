package brb

import (
	"crypto/sha256"
	"fmt"
	"math"
)

type BrachaImprovedMessage struct {
	BrachaMessage
	Included []uint64
}

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

	// Modification Bracha 2: Only a subset of nodes are participating in the agreement
	participatingEcho  map[brachaIdentifier]bool
	participatingReady map[brachaIdentifier]bool
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
	b.participatingEcho = make(map[brachaIdentifier]bool)
	b.participatingReady = make(map[brachaIdentifier]bool)

	if cfg.Byz {
		fmt.Printf("process %v is a Bracha Byzantine node\n", cfg.Id)
	}
}

func (b *BrachaImproved) send(messageType uint8, uid uint32, id brachaIdentifier, data interface{}, to []uint64) {
	// Only send one echo per message, and only if included
	if _, ok := b.echoSent[id]; messageType == BrachaEcho && (ok || !b.participatingEcho[id]) {
		return
	}

	// Only send one ready per message, and only if included
	if _, ok := b.readySent[id]; messageType == BrachaReady && (ok || !b.participatingReady[id]) {
		return
	}

	for _, n := range to {
		if n != b.cfg.Id {
			b.n.Send(messageType, n, uid, data)
		}
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

	m := data.(BrachaImprovedMessage)

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

	// Not yet known if participating
	if _, ok := b.participatingEcho[id]; !ok {
		// If included list is empty, not participating (Bracha modification 3?)
		if len(m.Included) == 0 {
			b.participatingEcho[id] = false
			b.participatingReady[id] = false
		}

		found := false
		for i, pid := range m.Included {
			if pid == b.cfg.Id {
				b.participatingEcho[id] = true
				b.participatingReady[id] = i <= b.cfg.F*2+1+b.cfg.F
				found = true
				break
			}
		}

		if !found {
			b.participatingEcho[id] = false
			b.participatingReady[id] = false
		}
	}

	del := b.hasDelivered(uid)
	switch messageType {
	case BrachaSend:
		// Modification Bracha 1: Use implicit echo messages
		b.send(BrachaEcho, uid, id, data, m.Included)

		b.echo[id][src] = struct{}{}
		b.echo[id][b.cfg.Id] = struct{}{}
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
	if len(b.echo[id]) >= int(math.Ceil((float64(b.cfg.N)+float64(b.cfg.F)+1)/2)) || len(b.ready[id]) >= b.cfg.F+1 {
		b.send(BrachaReady, uid, id, data, b.cfg.Neighbours)

		b.ready[id][b.cfg.Id] = struct{}{}
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

		b.participatingEcho[id] = true
		b.participatingReady[id] = true

		// TODO: use ID for now, switch to minimum cost later
		echoReq := int(math.Ceil((float64(b.cfg.N)+float64(b.cfg.F)+1)/2)) + b.cfg.F
		included := make([]uint64, 0, echoReq+1)

		included = append(included, b.cfg.Id)

		for _, pid := range b.cfg.Neighbours {
			included = append(included, pid)
			echoReq -= 1

			if echoReq == 0 {
				break
			}
		}

		m := BrachaImprovedMessage{
			BrachaMessage: BrachaMessage{
				Src:     b.cfg.Id,
				Payload: payload,
			},
			Included: included,
		}

		b.send(BrachaSend, uid, id, m, included)
	}
}
