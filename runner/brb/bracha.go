package brb

import (
	"crypto/sha256"
	"fmt"
	"math"
	"reflect"
	"rp-runner/graphs"
)

const (
	BrachaSend  uint8 = 1
	BrachaEcho  uint8 = 2
	BrachaReady uint8 = 3
)

type BrachaMessage struct {
	Src     uint64
	Id      uint32
	Payload Size
}

func (b BrachaMessage) SizeOf() uintptr {
	return reflect.TypeOf(b.Src).Size() + reflect.TypeOf(b.Id).Size() + b.Payload.SizeOf()
}

type brachaIdentifier struct {
	Src            uint64
	Id, TrackingId uint32
	Hash           [sha256.Size]byte
}

// Original Bracha Protocol
type Bracha struct {
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
}

var _ Protocol = (*Bracha)(nil)

func (b *Bracha) Init(n Network, app Application, cfg Config) {
	if cfg.Graph != nil && !graphs.IsFullyConnected(cfg.Graph) {
		panic("normal bracha does not work on non-fully connected networks!")
	}

	b.n = n
	b.app = app
	b.cfg = cfg
	b.delivered = make(map[brachaIdentifier]struct{})
	b.echo = make(map[brachaIdentifier]map[uint64]struct{})
	b.ready = make(map[brachaIdentifier]map[uint64]struct{})
	b.echoSent = make(map[brachaIdentifier]struct{})
	b.readySent = make(map[brachaIdentifier]struct{})

	if !cfg.Silent && cfg.Byz {
		fmt.Printf("process %v is a Bracha Byzantine node\n", cfg.Id)
	}
}

func (b *Bracha) send(messageType uint8, uid uint32, id brachaIdentifier, data BrachaMessage) {
	// Only send one echo per message
	if _, ok := b.echoSent[id]; ok && messageType == BrachaEcho {
		return
	}

	// Only send one ready per message
	if _, ok := b.readySent[id]; ok && messageType == BrachaReady {
		return
	}

	for _, n := range b.cfg.Neighbours {
		b.n.Send(messageType, n, uid, data, BroadcastInfo{
			Type: BrachaEveryone,
			Id:   b.bcId,
		})
	}

	b.bcId += 1
}

func (b *Bracha) hasDelivered(id brachaIdentifier) bool {
	_, ok := b.delivered[id]
	return ok
}

func (b *Bracha) Receive(messageType uint8, src uint64, uid uint32, data Size) {
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

	del := b.hasDelivered(id)
	switch messageType {
	case BrachaSend:
		b.send(BrachaEcho, uid, id, m)

		b.echoSent[id] = struct{}{}
		b.echo[id][b.cfg.Id] = struct{}{}
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
		b.send(BrachaReady, uid, id, m)

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

func (b *Bracha) Broadcast(uid uint32, payload Size, _ BroadcastInfo) {
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

		m := BrachaMessage{
			Src:     b.cfg.Id,
			Id:      b.cnt,
			Payload: payload,
		}

		b.send(BrachaSend, uid, id, m)
		b.send(BrachaEcho, uid, id, m)
		b.echoSent[id] = struct{}{}
		b.cnt += 1
	}
}

func (b *Bracha) Category() ProtocolCategory {
	return BrachaCat
}
