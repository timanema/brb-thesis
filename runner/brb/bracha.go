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
	BrachaSend uint8 = 1
	BrachaEcho uint8 = 2
	BrachaReady uint8 = 3
)

type BrachaMessage struct {
	Src uint64
	Payload []byte
}

// TODO: same as with Dolev, not really fair
type brachaIdentifier struct {
	Id uint32
	Hash [sha256.Size]byte
}

type Bracha struct {
	n Network
	app Application
	cfg Config

	delivered map[uint32]struct{}

	echo map[brachaIdentifier]map[uint64]struct{}
	ready map[brachaIdentifier]map[uint64]struct{}

	// TODO: check if uid is enough, or full id (including content hash) is needed
	echoSent map[uint32]struct{}
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

func (b *Bracha) send(messageType uint8, uid uint32, data []byte, ex uint64) {
	// Only send one echo per message
	if _, ok :=  b.echoSent[uid]; ok && messageType == BrachaEcho {
		return
	}

	// Only send one ready per message
	if _, ok :=  b.readySent[uid]; ok && messageType == BrachaReady {
		return
	}

	for _, n := range b.cfg.Neighbours {
		if n != ex{
			b.n.Send(messageType, n, uid, data)
		}
	}
}

func (b *Bracha) Receive(messageType uint8, src uint64, uid uint32, data []byte) {
	if b.cfg.Byz {
		// TODO: better byzantine behaviour?
		return
	}

	var m BrachaMessage
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	if err := dec.Decode(&m); err != nil {
		fmt.Printf("process %v errored while decoding bracha message: %v\n", d.cfg.Id, err)
		os.Exit(1)
	}

	id := brachaIdentifier{
		Id:   uid,
		Hash: sha256.Sum256(data),
	}

	if _, ok := b.echo[id]; !ok {
		b.echo[id] = make(map[uint64]struct{})
	}
	if _, ok := b.ready[id]; !ok {
		b.ready[id] = make(map[uint64]struct{})
	}

	switch messageType {
	case BrachaSend:
		b.send(BrachaEcho, uid, data, src)
		b.echoSent[uid] = struct{}{}
	case BrachaEcho:
		b.echo[id][src] = struct{}{}
	case BrachaReady:
		b.ready[id][src] = struct{}{}
	}

	// Send ready if enough ((n + f + 1) / 2) echos
	if len(b.echo[id]) > int(math.Ceil((float64(b.cfg.N) + float64(b.cfg.F) + 1) / 2)) {

	}
}

func (b *Bracha) Send(uid uint32, payload []byte) {

}
