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

type DolevImproved struct {
	n   Network
	app Application
	cfg Config

	delivered           map[uint32]struct{}
	paths               map[dolevIdentifier][]graphs.Path
	neighboursDelivered map[uint32]map[uint64]struct{}
}

func (d *DolevImproved) Init(n Network, app Application, cfg Config) {
	d.n = n
	d.app = app
	d.cfg = cfg
	d.delivered = make(map[uint32]struct{})
	d.paths = make(map[dolevIdentifier][]graphs.Path)
	d.neighboursDelivered = make(map[uint32]map[uint64]struct{})

	if cfg.Byz {
		fmt.Printf("process %v is a Dolev (improved) Byzantine node\n", cfg.Id)
	}
}

func (d *DolevImproved) send(uid uint32, data []byte, to []uint64) {
	for _, n := range to {
		d.n.Send(0, n, uid, data)
	}
}

func (d *DolevImproved) hasDelivered(uid uint32) bool {
	_, ok := d.delivered[uid]
	return ok
}

func (d *DolevImproved) deliver(uid uint32, payload []byte) {
	if !d.hasDelivered(uid) {
		d.delivered[uid] = struct{}{}
		d.app.Deliver(uid, payload)
	}
}

func (d *DolevImproved) Receive(_ uint8, src uint64, uid uint32, data []byte) {
	if d.cfg.Byz {
		// TODO: better byzantine behaviour?
		return
	}

	// Modification 5: Stop processing message once delivered
	if d.hasDelivered(uid) {
		return
	}

	var m DolevMessage
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	if err := dec.Decode(&m); err != nil {
		fmt.Printf("process %v errored while decoding dolev message: %v\n", d.cfg.Id, err)
		os.Exit(1)
	}

	if _, ok := d.neighboursDelivered[uid]; !ok {
		d.neighboursDelivered[uid] = make(map[uint64]struct{})
	}

	// Modification 2: A process has delivered the message when the path is empty
	if len(m.Path) == 0 {
		d.neighboursDelivered[uid][src] = struct{}{}
	}

	// Add latest edge to path for message
	m.Path = append(m.Path, simple.WeightedEdge{
		F: simple.Node(src),
		T: simple.Node(d.cfg.Id),
	})

	// Modification 1: Deliver when receiving from source
	if m.Src == src {
		d.deliver(uid, m.Payload)
	}

	// Add paths to mem for this message
	id := dolevIdentifier{
		Id:   uid,
		Hash: sha256.Sum256(m.Payload),
	}
	d.paths[id] = append(d.paths[id], m.Path)

	// Send to appropriate neighbours
	b := bytes.NewBuffer(make([]byte, 0, 20))
	enc := gob.NewEncoder(b)
	if err := enc.Encode(m); err != nil {
		fmt.Printf("process %v errored while encoding dolev message: %v\n", d.cfg.Id, err)
		os.Exit(1)
	}

	to := make([]uint64, 0, len(d.cfg.Neighbours))
	for _, n := range d.cfg.Neighbours {
		// Modification 4: No longer relay to neighbours who have delivered the message already
		if _, ok := d.neighboursDelivered[uid][n]; n != src && !pathContains(n, m.Path) && !ok {
			to = append(to, n)
		}
	}

	// Modification 2: If delivered, sent empty path
	if d.hasDelivered(uid) {
		m.Path = nil
	}

	d.send(uid, b.Bytes(), to)

	if _, ok := d.delivered[uid]; !ok {
		if graphs.VerifyDisjointPaths(d.paths[id], simple.Node(m.Src), simple.Node(d.cfg.Id), d.cfg.F+1) {
			d.delivered[uid] = struct{}{}
			d.app.Deliver(uid, m.Payload)
		}
	}
}

func (d *DolevImproved) Send(uid uint32, payload []byte) {
	if _, ok := d.delivered[uid]; !ok {
		id := dolevIdentifier{
			Id:   uid,
			Hash: sha256.Sum256(payload),
		}

		d.delivered[uid] = struct{}{}
		d.paths[id] = make([]graphs.Path, d.cfg.F*2+1)
		d.neighboursDelivered[uid] = make(map[uint64]struct{})
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
