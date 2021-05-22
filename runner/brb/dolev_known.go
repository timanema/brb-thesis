package brb

import (
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"os"
	"rp-runner/graphs"
	"strconv"
)

type dolevPath struct {
	Desired, Actual graphs.Path
}

type DolevKnownMessage struct {
	Src     uint64
	Id      uint32
	Payload interface{}
	Paths   dolevPath
}

// Dolev with routing for RP Tim Anema
type DolevKnown struct {
	n   Network
	app Application
	cfg Config

	cnt uint32

	delivered map[dolevIdentifier]struct{}
	paths     map[dolevIdentifier][]graphs.Path

	broadcast []graphs.Path
}

var _ Protocol = (*DolevKnown)(nil)

func (d *DolevKnown) Init(n Network, app Application, cfg Config) {
	d.n = n
	d.app = app
	d.cfg = cfg
	d.delivered = make(map[dolevIdentifier]struct{})
	d.paths = make(map[dolevIdentifier][]graphs.Path)

	if !cfg.Silent && cfg.Byz {
		fmt.Printf("process %v is a Dolev Byzantine node\n", cfg.Id)
		return
	}

	if d.broadcast == nil && !d.cfg.Unused {
		routes, err := graphs.BuildLookupTable(cfg.Graph, graphs.Node{
			Id:   int64(d.cfg.Id),
			Name: strconv.Itoa(int(d.cfg.Id)),
		}, d.cfg.F*2+1, 0, false)
		if err != nil {
			fmt.Printf("process %v errored while building lookup table: %v\n", d.cfg.Id, err)
			os.Exit(1)
		}
		d.broadcast = make([]graphs.Path, 0, len(routes))

		for _, g := range routes {
			d.broadcast = append(d.broadcast, g...)
		}

		if d.cfg.Id == 0 {
			//fmt.Println(d.routes)
			fmt.Println(d.cfg.Neighbours)
			//graphs.PrintGraphviz(graphs.Directed(cfg.Graph))
		}
	}
}

func (d *DolevKnown) sendMessage(uid uint32, m DolevKnownMessage) {
	if cur := len(m.Paths.Actual); len(m.Paths.Desired) > cur {
		path := make(graphs.Path, len(m.Paths.Actual))
		copy(path, m.Paths.Actual)
		m.Paths.Actual = path

		d.n.Send(0, uint64(m.Paths.Desired[cur].To().ID()), uid, m, BroadcastInfo{})
	}
}

func (d *DolevKnown) sendInitialMessage(uid uint32, payload interface{}) error {
	m := DolevKnownMessage{
		Src:     d.cfg.Id,
		Payload: payload,
	}

	for _, p := range d.broadcast {
		m.Paths = dolevPath{
			Desired: p,
			Actual:  nil,
		}

		d.n.Send(0, uint64(p[0].To().ID()), uid, m, BroadcastInfo{})
	}

	return nil
}

func (d *DolevKnown) hasDelivered(id dolevIdentifier) bool {
	_, ok := d.delivered[id]
	return ok
}

func (d *DolevKnown) Receive(_ uint8, src uint64, uid uint32, data interface{}) {
	if d.cfg.Byz {
		// TODO: better byzantine behaviour?
		return
	}

	m := data.(DolevKnownMessage)

	// Add paths to mem for this message
	id := dolevIdentifier{
		Src:        m.Src,
		Id:         m.Id,
		TrackingId: uid,
		Hash:       MustHash(m.Payload),
	}

	// Add latest edge to path for message
	m.Paths.Actual = append(m.Paths.Actual, simple.WeightedEdge{
		F: simple.Node(src),
		T: simple.Node(d.cfg.Id),
	})

	if !d.hasDelivered(id) {
		d.paths[id] = append(d.paths[id], m.Paths.Actual)
	}

	// Send to next hops
	d.sendMessage(uid, m)

	if !d.hasDelivered(id) {
		if graphs.VerifyDisjointPaths(d.paths[id], simple.Node(m.Src), simple.Node(d.cfg.Id), d.cfg.F+1) {
			d.delivered[id] = struct{}{}
			d.app.Deliver(uid, m.Payload, m.Src)

			// Memory cleanup
			delete(d.paths, id)
		}
	}
}

func (d *DolevKnown) Broadcast(uid uint32, payload interface{}) {
	id := dolevIdentifier{
		Src:        d.cfg.Id,
		Id:         d.cnt,
		TrackingId: uid,
		Hash:       MustHash(payload),
	}

	if _, ok := d.delivered[id]; !ok {
		d.delivered[id] = struct{}{}
		d.paths[id] = make([]graphs.Path, d.cfg.F*2+1)
		d.app.Deliver(uid, payload, 0)

		if err := d.sendInitialMessage(uid, payload); err != nil {
			fmt.Printf("process %v errored while broadcasting dolev (known) message: %v\n", d.cfg.Id, err)
			os.Exit(1)
		}

		d.cnt += 1
	}
}

func (d *DolevKnown) Category() ProtocolCategory {
	return DolevCat
}
