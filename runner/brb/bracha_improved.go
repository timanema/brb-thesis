package brb

import (
	"fmt"
	"math"
	"rp-runner/brb/algo"
	"rp-runner/graphs"
)

// Improved version for RP Tim Anema
type BrachaImproved struct {
	n   Network
	app Application
	cfg Config

	cnt  uint32
	bcId int

	delivered map[brachaIdentifier]struct{}

	echo  map[brachaIdentifier]map[uint64]struct{}
	ready map[brachaIdentifier]map[uint64]struct{}

	echoSent  map[brachaIdentifier]struct{}
	readySent map[brachaIdentifier]struct{}

	// Modification Bracha 2: Only a subset of nodes are participating in the agreement
	participatingEcho  map[brachaIdentifier]bool
	participatingReady map[brachaIdentifier]bool

	inclusion algo.BrachaInclusionTable
}

var _ Protocol = (*BrachaImproved)(nil)

func (b *BrachaImproved) Init(n Network, app Application, cfg Config) {
	b.n = n
	b.app = app
	b.cfg = cfg
	b.delivered = make(map[brachaIdentifier]struct{})
	b.echo = make(map[brachaIdentifier]map[uint64]struct{})
	b.ready = make(map[brachaIdentifier]map[uint64]struct{})
	b.echoSent = make(map[brachaIdentifier]struct{})
	b.readySent = make(map[brachaIdentifier]struct{})
	b.participatingEcho = make(map[brachaIdentifier]bool)
	b.participatingReady = make(map[brachaIdentifier]bool)

	c, bd := cfg.AdditionalConfig.(BrachaDolevConfig)

	if !cfg.Unused {
		if cfg.Graph == nil {
			panic("improved bracha needs graph!")
		}

		if !bd && !graphs.IsFullyConnected(cfg.Graph) {
			panic("improved bracha does not work on non-fully connected networks!")
		}
	}

	if bd {
		b.inclusion = c.Included
	} else {
		nodes, _ := graphs.Nodes(b.cfg.Graph)
		b.inclusion = algo.FindBrachaInclusionTable(b.cfg.Graph, nodes, b.cfg.N, b.cfg.F)
	}

	if !cfg.Silent && cfg.Byz {
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

	t := BrachaEveryone
	if b.cfg.OptimizationConfig.BrachaMinimalSubset && (messageType == BrachaSend || messageType == BrachaEcho) {
		t = BrachaPartial
	}

	i := b.bcId
	b.bcId += 1
	for _, n := range to {
		if n != b.cfg.Id {
			b.n.Send(messageType, n, uid, data, BroadcastInfo{
				Type: t,
				Id:   i,
			})
		}
	}
}

func (b *BrachaImproved) hasDelivered(id brachaIdentifier) bool {
	_, ok := b.delivered[id]
	return ok
}

func (b *BrachaImproved) Receive(messageType uint8, src uint64, uid uint32, data interface{}) {
	if b.cfg.Byz {
		// TODO: better byzantine behaviour?
		return
	}

	m := data.(BrachaMessage)

	id := brachaIdentifier{
		Src:        m.Src,
		Id:         m.Id,
		TrackingId: uid,
		Hash:       MustHash(m.Payload),
	}

	_, echoMade := b.echo[id]
	_, readyMade := b.ready[id]
	if !echoMade || !readyMade {
		b.echo[id] = make(map[uint64]struct{})
		b.ready[id] = make(map[uint64]struct{})
	}

	// Not yet known if participating
	if _, ok := b.participatingEcho[id]; !ok {
		found := false
		for i, pid := range b.inclusion[m.Src] {
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

	del := b.hasDelivered(id)
	switch messageType {
	case BrachaSend:
		// Modification Bracha 1: Use implicit echo messages
		b.send(BrachaEcho, uid, id, data, b.inclusion[m.Src])

		b.echo[id][b.cfg.Id] = struct{}{}
		b.echoSent[id] = struct{}{}

		if b.cfg.OptimizationConfig.BrachaImplicitEcho {
			b.echo[id][src] = struct{}{}
		}
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

		if !b.hasDelivered(id) {
			b.ready[id][b.cfg.Id] = struct{}{}
			b.readySent[id] = struct{}{}
		}
	}

	// Deliver if enough readys
	if !b.hasDelivered(id) && len(b.ready[id]) >= b.cfg.F*2+1 {
		b.delivered[id] = struct{}{}
		b.app.Deliver(uid, m.Payload, m.Src)

		// Memory cleanup
		delete(b.echo, id)
		delete(b.ready, id)
		delete(b.echoSent, id)
		delete(b.readySent, id)
	}
}

func (b *BrachaImproved) Broadcast(uid uint32, payload interface{}, _ BroadcastInfo) {
	id := brachaIdentifier{
		Src:        b.cfg.Id,
		Id:         b.cnt,
		TrackingId: uid,
		Hash:       MustHash(payload),
	}

	if _, ok := b.delivered[id]; !ok {
		b.echo[id] = map[uint64]struct{}{
			b.cfg.Id: {},
		}
		b.ready[id] = make(map[uint64]struct{})

		b.participatingEcho[id] = true
		b.participatingReady[id] = true

		m := BrachaMessage{
			Src:     b.cfg.Id,
			Id:      b.cnt,
			Payload: payload,
		}
		b.cnt += 1

		b.send(BrachaSend, uid, id, m, b.inclusion[b.cfg.Id])

		if !b.cfg.OptimizationConfig.BrachaImplicitEcho {
			b.send(BrachaEcho, uid, id, m, b.inclusion[b.cfg.Id])
		}

		b.echoSent[id] = struct{}{}

	}
}

func (b *BrachaImproved) Category() ProtocolCategory {
	return BrachaCat
}
