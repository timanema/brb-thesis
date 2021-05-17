package brb

import (
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"rp-runner/graphs"
)

// Improved Dolev Protocol based on Bonomi (Multi-hop)
type DolevImproved struct {
	n   Network
	app Application
	cfg Config

	cnt uint32

	delivered           map[dolevIdentifier]struct{}
	paths               map[dolevIdentifier][]graphs.Path
	neighboursDelivered map[dolevIdentifier]map[uint64]struct{}
}

var _ Protocol = (*DolevImproved)(nil)

func (d *DolevImproved) Init(n Network, app Application, cfg Config) {
	d.n = n
	d.app = app
	d.cfg = cfg
	d.delivered = make(map[dolevIdentifier]struct{})
	d.paths = make(map[dolevIdentifier][]graphs.Path)
	d.neighboursDelivered = make(map[dolevIdentifier]map[uint64]struct{})

	if !cfg.Silent && cfg.Byz {
		fmt.Printf("process %v is a Dolev (improved) Byzantine node\n", cfg.Id)
	}
}

func (d *DolevImproved) send(uid uint32, data interface{}, to []uint64) {
	for _, n := range to {
		d.n.Send(0, n, uid, data)
	}
}

func (d *DolevImproved) hasDelivered(id dolevIdentifier) bool {
	_, ok := d.delivered[id]
	return ok
}

func (d *DolevImproved) deliver(uid uint32, id dolevIdentifier, m DolevMessage) {
	if !d.hasDelivered(id) {
		d.delivered[id] = struct{}{}
		d.app.Deliver(uid, m.Payload, m.Src)

		// Memory cleanup
		delete(d.paths, id)
		delete(d.neighboursDelivered, id)
	}
}

func (d *DolevImproved) Receive(_ uint8, src uint64, uid uint32, data interface{}) {
	if d.cfg.Byz {
		// TODO: better byzantine behaviour?
		return
	}

	m := data.(DolevMessage)
	id := dolevIdentifier{
		Src:        m.Src,
		Id:         m.Id,
		TrackingId: uid,
		Hash:       MustHash(m.Payload),
	}

	// Modification 5: Stop processing message once delivered
	if d.hasDelivered(id) {
		return
	}

	if _, ok := d.neighboursDelivered[id]; !ok {
		d.neighboursDelivered[id] = make(map[uint64]struct{})
	}

	traversed := make(map[uint64]struct{}, len(m.Path))

	// Modification 4: Stop relaying messages which contain the label of nodes that already delivered
	for _, e := range m.Path {
		traversed[uint64(e.From().ID())] = struct{}{}

		if _, ok := d.neighboursDelivered[id][uint64(e.From().ID())]; ok {
			return
		}
	}

	// Modification 2: A process has delivered the message when the path is empty
	if len(m.Path) == 0 {
		d.neighboursDelivered[id][src] = struct{}{}

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
		d.deliver(uid, id, m)
	}

	if d.cfg.Id == m.Src {
		panic("received message from self, should have been delivered already")
	}

	if _, ok := d.delivered[id]; !ok {
		// Add paths to mem for this message
		d.paths[id] = append(d.paths[id], m.Path)

		if graphs.VerifyDisjointPaths(d.paths[id], simple.Node(m.Src), simple.Node(d.cfg.Id), d.cfg.F+1) {
			d.deliver(uid, id, m)
		} else {
			//fmt.Printf("proc %v is NOT delivering %+v\n", d.cfg.Id, id)
		}
	}

	// Modification 2: If delivered, sent empty path
	del := d.hasDelivered(id)
	if del {
		m.Path = nil
	}

	// Send to appropriate neighbours
	to := make([]uint64, 0, len(d.cfg.Neighbours))

	for _, n := range d.cfg.Neighbours {
		_, trav := traversed[n]
		// Modification 3: No longer relay to neighbours who have delivered the message already
		if _, ok := d.neighboursDelivered[id][n]; n != src && !ok && (del || !trav) {
			to = append(to, n)
		}
	}

	d.send(uid, m, to)
}

func (d *DolevImproved) Broadcast(uid uint32, payload interface{}) {
	id := dolevIdentifier{
		Src:        d.cfg.Id,
		Id:         d.cnt,
		TrackingId: uid,
		Hash:       MustHash(payload),
	}

	if _, ok := d.delivered[id]; !ok {
		d.delivered[id] = struct{}{}
		d.paths[id] = make([]graphs.Path, d.cfg.F*2+1)
		d.neighboursDelivered[id] = make(map[uint64]struct{})
		d.app.Deliver(uid, payload, 0)

		m := DolevMessage{
			Src:     d.cfg.Id,
			Id:      d.cnt,
			Path:    nil,
			Payload: payload,
		}

		d.send(uid, m, d.cfg.Neighbours)
		d.cnt += 1
	}
}
