package brb

import (
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"os"
	"reflect"
	"rp-runner/brb/algo"
	"rp-runner/graphs"
	"strconv"
)

type DolevKnownMessage struct {
	Src     uint64
	Id      uint32
	Payload Size
	Path    algo.DolevPath
}

func (d DolevKnownMessage) SizeOf() uintptr {
	return reflect.TypeOf(d.Src).Size() + reflect.TypeOf(d.Id).Size() + d.Payload.SizeOf() + d.Path.SizeOf()
}

// Dolev with routing for RP Tim Anema
type DolevKnown struct {
	n   Network
	app Application
	cfg Config

	cnt uint32

	delivered map[dolevIdentifier]struct{}
	paths     map[dolevIdentifier][]graphs.Path

	broadcast algo.BroadcastPlan
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
		routes, err := algo.BuildRoutingTable(cfg.Graph, graphs.Node{
			Id:   int64(d.cfg.Id),
			Name: strconv.Itoa(int(d.cfg.Id)),
		}, d.cfg.F*2+1, 0, false)
		if err != nil {
			panic(fmt.Sprintf("process %v errored while building lookup table: %v\n", d.cfg.Id, err))
		}

		d.broadcast = algo.DolevRouting(routes, false, false)
	}
}

func (d *DolevKnown) sendMessage(uid uint32, m DolevKnownMessage) {
	if cur := len(m.Path.Actual); len(m.Path.Desired) > cur {
		path := make(graphs.Path, len(m.Path.Actual))
		copy(path, m.Path.Actual)
		m.Path.Actual = path

		d.n.TriggerStat(uid, StartRelay)
		d.n.Send(0, uint64(m.Path.Desired[cur].To().ID()), uid, m, BroadcastInfo{})
	}
}

func (d *DolevKnown) sendInitialMessage(uid uint32, payload Size) error {
	m := DolevKnownMessage{
		Src:     d.cfg.Id,
		Payload: payload,
	}

	for dst, paths := range d.broadcast {
		for _, p := range paths {
			m.Path = algo.DolevPath{
				Desired: p.P,
				Prio:    p.Prio,
			}

			d.n.Send(0, dst, uid, m, BroadcastInfo{})
		}
	}

	return nil
}

func (d *DolevKnown) hasDelivered(id dolevIdentifier) bool {
	_, ok := d.delivered[id]
	return ok
}

func (d *DolevKnown) Receive(_ uint8, src uint64, uid uint32, data Size) {
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
	m.Path.Actual = append(m.Path.Actual, simple.WeightedEdge{
		F: simple.Node(src),
		T: simple.Node(d.cfg.Id),
	})

	if !d.hasDelivered(id) {
		d.paths[id] = append(d.paths[id], m.Path.Actual)
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

func (d *DolevKnown) Broadcast(uid uint32, payload Size, _ BroadcastInfo) {
	id := dolevIdentifier{
		Src:        d.cfg.Id,
		Id:         d.cnt,
		TrackingId: uid,
		Hash:       MustHash(payload),
	}

	if _, ok := d.delivered[id]; !ok {
		d.delivered[id] = struct{}{}
		d.paths[id] = make([]graphs.Path, d.cfg.F*2+1)
		d.app.Deliver(uid, payload, d.cfg.Id)

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
