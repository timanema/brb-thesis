package brb

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph/simple"
	"os"
	"rp-runner/graphs"
	"strconv"
)

type DolevKnownImproved struct {
	n   Network
	app Application
	cfg Config

	delivered map[uint32]struct{}
	paths     map[dolevIdentifier][]graphs.Path
	routes    map[uint64][]graphs.Path
	broadcast []graphs.Path
}

func (d *DolevKnownImproved) Init(n Network, app Application, cfg Config) {
	d.n = n
	d.app = app
	d.cfg = cfg
	d.delivered = make(map[uint32]struct{})
	d.paths = make(map[dolevIdentifier][]graphs.Path)

	if d.routes == nil {
		routes, err := graphs.BuildLookupTable(cfg.Graph, graphs.Node{
			Id:   int64(d.cfg.Id),
			Name: strconv.Itoa(int(d.cfg.Id)),
		}, d.cfg.F*2+1, true)
		if err != nil {
			fmt.Printf("process %v errored while building lookup table: %v\n", d.cfg.Id, err)
			os.Exit(1)
		}
		d.routes = routes
		d.broadcast = make([]graphs.Path, 0, len(d.routes))

		for _, g := range d.routes {
			d.broadcast = append(d.broadcast, g...)
		}
	}

	if cfg.Byz {
		fmt.Printf("process %v is a Dolev Byzantine node\n", cfg.Id)
	}
}

func (d *DolevKnownImproved) sendMergedMessage(uid uint32, m DolevKnownMessage) error {
	next := combineDolevPaths(m.Paths)

	for dst, p := range next {
		b := bytes.NewBuffer(make([]byte, 0, 20))
		enc := gob.NewEncoder(b)

		m.Paths = p
		if err := enc.Encode(m); err != nil {
			return errors.Wrapf(err, "process %v errored while encoding dolev (known, improved) message", d.cfg.Id)
		}

		d.n.Send(0, dst, uid, b.Bytes())
	}

	return nil
}

func (d *DolevKnownImproved) sendInitialMessage(uid uint32, payload []byte) error {
	next := combinePaths(d.broadcast)

	m := &DolevKnownMessage{
		Src:     d.cfg.Id,
		Payload: payload,
	}

	for dst, p := range next {
		b := bytes.NewBuffer(make([]byte, 0, 20))
		enc := gob.NewEncoder(b)

		m.Paths = p
		if err := enc.Encode(m); err != nil {
			return errors.Wrapf(err, "process %v errored while encoding dolev (known, improved) message", d.cfg.Id)
		}

		d.n.Send(0, dst, uid, b.Bytes())
	}

	return nil
}

func (d *DolevKnownImproved) hasDelivered(uid uint32) bool {
	_, ok := d.delivered[uid]
	return ok
}

func (d *DolevKnownImproved) Receive(_ uint8, src uint64, uid uint32, data []byte) {
	if d.cfg.Byz {
		// TODO: better byzantine behaviour?
		return
	}

	var m DolevKnownMessage
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	if err := dec.Decode(&m); err != nil {
		fmt.Printf("process %v errored while decoding dolev (known, improved) message: %v\n", d.cfg.Id, err)
		os.Exit(1)
	}

	// Add paths to mem for this message
	id := dolevIdentifier{
		Id:   uid,
		Hash: sha256.Sum256(m.Payload),
	}

	// Add latest edge to path for message
	for i, p := range m.Paths {
		p.Actual = append(p.Actual, simple.WeightedEdge{
			F: simple.Node(src),
			T: simple.Node(d.cfg.Id),
		})
		m.Paths[i] = p

		if !d.hasDelivered(uid) {
			d.paths[id] = append(d.paths[id], p.Actual)
		}
	}

	// Send to next hops
	if err := d.sendMergedMessage(uid, m); err != nil {
		fmt.Printf("process %v errored while sending dolev (known, improved) message: %v\n", d.cfg.Id, err)
		os.Exit(1)
	}

	if !d.hasDelivered(uid) {
		// Additional modification (based on bonomi 7): Accept messages from origin immediately
		if m.Src == src || graphs.VerifyDisjointPaths(d.paths[id], simple.Node(m.Src), simple.Node(d.cfg.Id), d.cfg.F+1) {
			//fmt.Printf("proc %v is delivering %v at %v\n", d.cfg.Id, id, time.Now())
			d.delivered[uid] = struct{}{}
			d.app.Deliver(uid, m.Payload)

			// Memory cleanup
			delete(d.paths, id)
		}
	}
}

func (d *DolevKnownImproved) Send(uid uint32, payload []byte) {
	if _, ok := d.delivered[uid]; !ok {
		id := dolevIdentifier{
			Id:   uid,
			Hash: sha256.Sum256(payload),
		}

		d.delivered[uid] = struct{}{}
		d.paths[id] = make([]graphs.Path, d.cfg.F*2+1)
		d.app.Deliver(uid, payload)

		if err := d.sendInitialMessage(uid, payload); err != nil {
			fmt.Printf("process %v errored while broadcasting dolev (known, improved) message: %v\n", d.cfg.Id, err)
			os.Exit(1)
		}
	}
}
