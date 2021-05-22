package brb

import (
	"crypto/sha256"
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"rp-runner/graphs"
)

type DolevMessage struct {
	Src     uint64
	Id      uint32
	Path    graphs.Path
	Payload interface{}
}

type dolevIdentifier struct {
	Src            uint64
	Id, TrackingId uint32
	Hash           [sha256.Size]byte
}

// Original Dolev Protocol
type Dolev struct {
	n   Network
	app Application
	cfg Config

	cnt uint32

	delivered map[dolevIdentifier]struct{}
	paths     map[dolevIdentifier][]graphs.Path
}

var _ Protocol = (*Dolev)(nil)

func (d *Dolev) Init(n Network, app Application, cfg Config) {
	d.n = n
	d.app = app
	d.cfg = cfg
	d.delivered = make(map[dolevIdentifier]struct{})
	d.paths = make(map[dolevIdentifier][]graphs.Path)

	if !cfg.Silent && cfg.Byz {
		fmt.Printf("process %v is a Dolev Byzantine node\n", cfg.Id)
	}
}

func (d *Dolev) send(uid uint32, m DolevMessage, to []uint64) {
	for _, n := range to {
		path := make(graphs.Path, len(m.Path))
		copy(path, m.Path)
		m.Path = path
		d.n.Send(0, n, uid, m, BroadcastInfo{})
	}
}

func (d *Dolev) Receive(_ uint8, src uint64, uid uint32, data interface{}) {
	if d.cfg.Byz {
		// TODO: better byzantine behaviour?
		return
	}

	m := data.(DolevMessage)

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
		Src:        m.Src,
		Id:         m.Id,
		TrackingId: uid,
		Hash:       MustHash(m.Payload),
	}

	// Send to neighbours (except origin)
	to := make([]uint64, 0, len(d.cfg.Neighbours))
	for _, n := range d.cfg.Neighbours {
		if _, ok := traversed[n]; n != src && !ok {
			to = append(to, n)
		}
	}

	if _, ok := d.delivered[id]; !ok {
		d.paths[id] = append(d.paths[id], m.Path)

		if graphs.VerifyDisjointPaths(d.paths[id], simple.Node(m.Src), simple.Node(d.cfg.Id), d.cfg.F+1) {
			//fmt.Printf("proc %v is delivering %v at %v\n", d.cfg.Id, id, time.Now())
			d.delivered[id] = struct{}{}
			d.app.Deliver(uid, m.Payload, m.Src)

			// Memory cleanup
			delete(d.paths, id)
		}
	}

	d.send(uid, m, to)
}

func (d *Dolev) Broadcast(uid uint32, payload interface{}) {
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

func (d *Dolev) Category() ProtocolCategory {
	return DolevCat
}
