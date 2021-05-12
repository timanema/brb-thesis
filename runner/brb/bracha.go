package brb

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math"
	"os"
)

const (
	BrachaSend  uint8 = 1
	BrachaEcho  uint8 = 2
	BrachaReady uint8 = 3
)

type BrachaMessage struct {
	Src     uint64
	Payload []byte
}

// TODO: same as with Dolev, not really fair
type brachaIdentifier struct {
	Id   uint32
	Hash [sha256.Size]byte
}

type Bracha struct {
	n   Network
	app Application
	cfg Config

	delivered map[uint32]struct{}

	echo  map[brachaIdentifier]map[uint64]struct{}
	ready map[brachaIdentifier]map[uint64]struct{}

	// TODO: check if uid is enough, or full id (including content hash) is needed
	echoSent  map[uint32]struct{}
	readySent map[uint32]struct{}
}

func (b *Bracha) Init(n Network, app Application, cfg Config) {
	b.n = n
	b.app = app
	b.cfg = cfg
	b.delivered = make(map[uint32]struct{})
	b.echo = make(map[brachaIdentifier]map[uint64]struct{})
	b.ready = make(map[brachaIdentifier]map[uint64]struct{})
	b.echoSent = make(map[uint32]struct{})
	b.readySent = make(map[uint32]struct{})

	if cfg.Byz {
		fmt.Printf("process %v is a Bracha Byzantine node\n", cfg.Id)
	}
}

func (b *Bracha) send(messageType uint8, uid uint32, data interface{}) {
	// Only send one echo per message
	if _, ok := b.echoSent[uid]; ok && messageType == BrachaEcho {
		return
	}

	// Only send one ready per message
	if _, ok := b.readySent[uid]; ok && messageType == BrachaReady {
		return
	}

	for _, n := range b.cfg.Neighbours {
		b.n.Send(messageType, n, uid, data)
	}
}

func (b *Bracha) hasDelivered(uid uint32) bool {
	_, ok := b.delivered[uid]
	return ok
}

func (b *Bracha) Receive(messageType uint8, src uint64, uid uint32, data interface{}) {
	if b.cfg.Byz {
		// TODO: better byzantine behaviour?
		return
	}

	m := data.(BrachaMessage)
	//dec := gob.NewDecoder(bytes.NewBuffer(data))
	//if err := dec.Decode(&m); err != nil {
	//	fmt.Printf("process %v errored while decoding bracha message: %v\n", b.cfg.Id, err)
	//	os.Exit(1)
	//}

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
		b.send(BrachaEcho, uid, data)
		b.echoSent[uid] = struct{}{}
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
		b.send(BrachaReady, uid, data)
		b.readySent[uid] = struct{}{}
	}

	// Deliver if enough readys
	if !b.hasDelivered(uid) && len(b.ready[id]) >= b.cfg.F*2+1 {
		b.delivered[uid] = struct{}{}
		b.app.Deliver(uid, m.Payload)

		// Memory cleanup
		delete(b.echo, id)
		delete(b.ready, id)
		delete(b.echoSent, uid)
		delete(b.readySent, uid)
	}
}

func (b *Bracha) Send(uid uint32, payload []byte) {
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

		m := &BrachaMessage{
			Src:     b.cfg.Id,
			Payload: payload,
		}

		buf := bytes.NewBuffer(make([]byte, 0, 20))
		enc := gob.NewEncoder(buf)
		if err := enc.Encode(m); err != nil {
			fmt.Printf("process %v errored while encoding bracha message: %v\n", b.cfg.Id, err)
			os.Exit(1)
		}

		b.send(BrachaSend, uid, buf.Bytes())
	}
}
