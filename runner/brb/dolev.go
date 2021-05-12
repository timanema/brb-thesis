package brb

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"os"
	"rp-runner/graphs"
)

type DolevMessage struct {
	Src     uint64
	Path    graphs.Path
	Payload []byte
}

// TODO: cheating by using statistics tracking uid as dolev id, should probably change before using eval
type dolevIdentifier struct {
	Id   uint32
	Hash [sha256.Size]byte
}

type Dolev struct {
	n   Network
	app Application
	cfg Config

	delivered map[uint32]struct{}
	paths     map[dolevIdentifier][]graphs.Path
}

func (d *Dolev) Init(n Network, app Application, cfg Config) {
	d.n = n
	d.app = app
	d.cfg = cfg
	d.delivered = make(map[uint32]struct{})
	d.paths = make(map[dolevIdentifier][]graphs.Path)

	if cfg.Byz {
		fmt.Printf("process %v is a Dolev Byzantine node\n", cfg.Id)
	}
}

func (d *Dolev) send(uid uint32, data []byte, to []uint64) {
	for _, n := range to {
		d.n.Send(0, n, uid, data)
	}
}

func (d *Dolev) Receive(_ uint8, src uint64, uid uint32, data []byte) {
	if d.cfg.Byz {
		// TODO: better byzantine behaviour?
		return
	}

	var m DolevMessage
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	if err := dec.Decode(&m); err != nil {
		fmt.Printf("process %v errored while decoding dolev message: %v\n", d.cfg.Id, err)
		os.Exit(1)
	}

	traversed := make(map[uint64]struct{}, len(m.Path))
	for _, e := range m.Path {
		traversed[uint64(e.From().ID())] = struct{}{}
	}

	// Add latest edge to path for message
	m.Path = append(m.Path, simple.WeightedEdge{
		F: simple.Node(src),
		T: simple.Node(d.cfg.Id),
	})

	// Add paths to mem for this message
	id := dolevIdentifier{
		Id:   uid,
		Hash: sha256.Sum256(m.Payload),
	}

	// Send to neighbours (except origin)
	b := bytes.NewBuffer(make([]byte, 0, 20))
	enc := gob.NewEncoder(b)
	if err := enc.Encode(m); err != nil {
		fmt.Printf("process %v errored while encoding dolev message: %v\n", d.cfg.Id, err)
		os.Exit(1)
	}

	to := make([]uint64, 0, len(d.cfg.Neighbours))
	for _, n := range d.cfg.Neighbours {
		if _, ok := traversed[n]; n != src && !ok {
			to = append(to, n)
		}
	}

	if _, ok := d.delivered[uid]; !ok {
		d.paths[id] = append(d.paths[id], m.Path)

		if graphs.VerifyDisjointPaths(d.paths[id], simple.Node(m.Src), simple.Node(d.cfg.Id), d.cfg.F+1) {
			//fmt.Printf("proc %v is delivering %v at %v\n", d.cfg.Id, id, time.Now())
			d.delivered[uid] = struct{}{}
			d.app.Deliver(uid, m.Payload)

			// Memory cleanup
			delete(d.paths, id)
		}
	}

	d.send(uid, b.Bytes(), to)
}

func (d *Dolev) Send(uid uint32, payload []byte) {
	if _, ok := d.delivered[uid]; !ok {
		id := dolevIdentifier{
			Id:   uid,
			Hash: sha256.Sum256(payload),
		}

		d.delivered[uid] = struct{}{}
		d.paths[id] = make([]graphs.Path, d.cfg.F*2+1)
		d.app.Deliver(uid, payload)

		m := &DolevMessage{
			Src:     d.cfg.Id,
			Path:    nil,
			Payload: payload,
		}

		b := bytes.NewBuffer(make([]byte, 0, 20))
		enc := gob.NewEncoder(b)
		if err := enc.Encode(m); err != nil {
			fmt.Printf("process %v errored while encoding dolev message: %v\n", d.cfg.Id, err)
			os.Exit(1)
		}

		d.send(uid, b.Bytes(), d.cfg.Neighbours)
	}
}
