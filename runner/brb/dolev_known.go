package brb

import (
	"crypto/sha256"
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
	Payload []byte
	Paths   []dolevPath
}

// Dolev with routing for RP Tim Anema
type DolevKnown struct {
	n   Network
	app Application
	cfg Config

	delivered map[uint32]struct{}
	paths     map[dolevIdentifier][]graphs.Path
	routes    map[uint64][]graphs.Path
	broadcast []graphs.Path
}

func (d *DolevKnown) Init(n Network, app Application, cfg Config) {
	d.n = n
	d.app = app
	d.cfg = cfg
	d.delivered = make(map[uint32]struct{})
	d.paths = make(map[dolevIdentifier][]graphs.Path)

	if d.routes == nil {
		routes, err := graphs.BuildLookupTable(cfg.Graph, graphs.Node{
			Id:   int64(d.cfg.Id),
			Name: strconv.Itoa(int(d.cfg.Id)),
		}, d.cfg.F*2+1, false)
		if err != nil {
			fmt.Printf("process %v errored while building lookup table: %v\n", d.cfg.Id, err)
			os.Exit(1)
		}
		d.routes = routes
		d.broadcast = make([]graphs.Path, 0, len(d.routes))

		for _, g := range d.routes {
			d.broadcast = append(d.broadcast, g...)
		}

		if d.cfg.Id == 0 {
			//fmt.Println(d.routes)
			fmt.Println(d.cfg.Neighbours)
			//graphs.PrintGraphviz(graphs.Directed(cfg.Graph))
		}
	}

	if cfg.Byz {
		fmt.Printf("process %v is a Dolev Byzantine node\n", cfg.Id)
	}
}

func combinePaths(paths []graphs.Path) map[uint64][]dolevPath {
	res := make(map[uint64][]dolevPath)

	for _, p := range paths {
		next := uint64(p[0].To().ID())
		res[next] = append(res[next], dolevPath{Desired: p})
	}

	return res
}

func combineDolevPaths(paths []dolevPath) map[uint64][]dolevPath {
	res := make(map[uint64][]dolevPath)

	for _, p := range paths {
		if cur := len(p.Actual); len(p.Desired) > cur {
			next := uint64(p.Desired[cur].To().ID())
			res[next] = append(res[next], p)
		}
	}

	return res
}

func (d *DolevKnown) sendMergedMessage(uid uint32, m DolevKnownMessage) error {
	next := combineDolevPaths(m.Paths)

	for dst, p := range next {
		m.Paths = p

		d.n.Send(0, dst, uid, m)
	}

	return nil
}

func (d *DolevKnown) sendInitialMessage(uid uint32, payload []byte) error {
	next := combinePaths(d.broadcast)

	m := DolevKnownMessage{
		Src:     d.cfg.Id,
		Payload: payload,
	}

	for dst, p := range next {
		m.Paths = p

		d.n.Send(0, dst, uid, m)
	}

	return nil
}

func (d *DolevKnown) hasDelivered(uid uint32) bool {
	_, ok := d.delivered[uid]
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
		fmt.Printf("process %v errored while sending dolev (known) message: %v\n", d.cfg.Id, err)
		os.Exit(1)
	}

	if !d.hasDelivered(uid) {
		if graphs.VerifyDisjointPaths(d.paths[id], simple.Node(m.Src), simple.Node(d.cfg.Id), d.cfg.F+1) {
			d.delivered[uid] = struct{}{}
			d.app.Deliver(uid, m.Payload)

			// Memory cleanup
			delete(d.paths, id)
		}
	}
}

func (d *DolevKnown) Send(uid uint32, payload []byte) {
	if _, ok := d.delivered[uid]; !ok {
		id := dolevIdentifier{
			Id:   uid,
			Hash: sha256.Sum256(payload),
		}

		d.delivered[uid] = struct{}{}
		d.paths[id] = make([]graphs.Path, d.cfg.F*2+1)
		d.app.Deliver(uid, payload)

		if err := d.sendInitialMessage(uid, payload); err != nil {
			fmt.Printf("process %v errored while broadcasting dolev (known) message: %v\n", d.cfg.Id, err)
			os.Exit(1)
		}
	}
}
