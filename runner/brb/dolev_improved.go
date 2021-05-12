package brb

import (
	"crypto/sha256"
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"rp-runner/graphs"
)

// Improved Dolev Protocol based on Bonomi (Multi-hop)
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

func (d *DolevImproved) send(uid uint32, data interface{}, to []uint64) {
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

func (d *DolevImproved) Receive(_ uint8, src uint64, uid uint32, data interface{}) {
	if d.cfg.Byz {
		// TODO: better byzantine behaviour?
		return
	}

	// Modification 5: Stop processing message once delivered
	if d.hasDelivered(uid) {
		return
	}

	m := data.(DolevMessage)

	if _, ok := d.neighboursDelivered[uid]; !ok {
		d.neighboursDelivered[uid] = make(map[uint64]struct{})
	}

	traversed := make(map[uint64]struct{}, len(m.Path))

	// Modification 4: Stop relaying messages which contain the label of nodes that already delivered
	for _, e := range m.Path {
		traversed[uint64(e.From().ID())] = struct{}{}

		if _, ok := d.neighboursDelivered[uid][uint64(e.From().ID())]; ok {
			return
		}
	}

	// Modification 2: A process has delivered the message when the path is empty
	if len(m.Path) == 0 {
		d.neighboursDelivered[uid][src] = struct{}{}

		// Since the source process has delivered the message, there must be a link (either direct or over f+1 paths)
		if src != m.Src {
			m.Path = graphs.Path{
				simple.WeightedEdge{
					F: simple.Node(m.Src),
					T: simple.Node(src),
				},
			}
		}
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

	if _, ok := d.delivered[uid]; !ok {
		d.paths[id] = append(d.paths[id], m.Path)

		if graphs.VerifyDisjointPaths(d.paths[id], simple.Node(m.Src), simple.Node(d.cfg.Id), d.cfg.F+1) {
			d.delivered[uid] = struct{}{}
			d.app.Deliver(uid, m.Payload)

			// Memory cleanup
			delete(d.paths, id)
			delete(d.neighboursDelivered, uid)
		}
	}

	// Modification 2: If delivered, sent empty path
	del := d.hasDelivered(uid)
	if del {
		m.Path = nil
	}

	// Send to appropriate neighbours
	to := make([]uint64, 0, len(d.cfg.Neighbours))

	for _, n := range d.cfg.Neighbours {
		_, trav := traversed[n]
		// Modification 3: No longer relay to neighbours who have delivered the message already
		if _, ok := d.neighboursDelivered[uid][n]; n != src && !ok && (del || !trav) {
			to = append(to, n)
		}
	}

	d.send(uid, m, to)
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

		m := DolevMessage{
			Src:     d.cfg.Id,
			Path:    nil,
			Payload: payload,
		}

		d.send(uid, m, d.cfg.Neighbours)
	}
}
