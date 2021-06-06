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

type DolevKnownImprovedMessage struct {
	Src     uint64
	Id      uint32
	Payload Size
	Paths   []algo.DolevPath
}

func (d DolevKnownImprovedMessage) SizeOf() uintptr {
	return reflect.TypeOf(d.Src).Size() + reflect.TypeOf(d.Id).Size() + d.Payload.SizeOf() + algo.SizeOfMultiplePaths(d.Paths)
}

// Dolev with routing and additional optimizations for RP Tim Anema
type DolevKnownImproved struct {
	n   Network
	app Application
	cfg Config

	cnt uint32

	delivered map[dolevIdentifier]struct{}
	paths     map[dolevIdentifier][]graphs.Path

	buffer map[dolevIdentifier][]algo.DolevPath

	broadcast algo.BroadcastPlan

	bd     bool
	bdPlan map[uint64]algo.BroadcastPlan
}

var _ Protocol = (*DolevKnownImproved)(nil)

func (d *DolevKnownImproved) Init(n Network, app Application, cfg Config) {
	d.n = n
	d.app = app
	d.cfg = cfg
	d.delivered = make(map[dolevIdentifier]struct{})
	d.paths = make(map[dolevIdentifier][]graphs.Path)
	d.buffer = make(map[dolevIdentifier][]algo.DolevPath)

	if !cfg.Silent && cfg.Byz {
		fmt.Printf("process %v is a Dolev (known improved) Byzantine node\n", cfg.Id)
		return
	}

	if d.broadcast == nil && !d.cfg.Unused {
		_, bd := cfg.AdditionalConfig.(BrachaDolevConfig)
		d.bd = bd

		routes, err := algo.BuildRoutingTable(cfg.Graph, graphs.Node{
			Id:   int64(d.cfg.Id),
			Name: strconv.Itoa(int(d.cfg.Id)),
		}, d.cfg.F*2+1, 5, true)
		if err != nil {
			panic(fmt.Sprintf("process %v errored while building lookup table: %v\n", d.cfg.Id, err))
		}

		algo.FixDeadlocks(routes)

		if d.bd {
			n, m := graphs.Nodes(d.cfg.Graph)
			d.bdPlan = algo.BrachaDolevRouting(routes, graphs.FindAdjMap(graphs.Directed(d.cfg.Graph), m), n, d.cfg.N, d.cfg.F)
		}

		d.broadcast = algo.DolevRouting(routes, true, true)
	}
}

func (d *DolevKnownImproved) sendMergedMessage(uid uint32, id dolevIdentifier, m DolevKnownImprovedMessage) error {
	del := d.hasDelivered(id)

	paths := make([]algo.DolevPath, 0, len(m.Paths))

	// If delivered, relay all messages, including ones in the buffer
	if del {
		paths = append(paths, m.Paths...)

		if buf, ok := d.buffer[id]; ok {
			paths = append(paths, buf...)
		}

		// Clear buffer
		d.buffer[id] = nil
	} else {
		// If not delivered, add all messages with no priority to the buffer and
		// all messages with priority to the outgoing paths
		for _, p := range m.Paths {
			if len(p.Desired) == len(p.Actual) {
				continue
			}

			if p.Prio {
				paths = append(paths, p)
			} else {
				d.buffer[id] = append(d.buffer[id], p)
			}
		}
	}

	// TODO: let paths with the same next hop join if a message is traveling there

	next := algo.CombineDolevPaths(paths)

	if len(next) > 0 {
		d.n.TriggerStat(uid, StartRelay)
	}

	for dst, p := range next {
		m.Paths = p
		d.n.Send(0, dst, uid, m, BroadcastInfo{})
	}

	return nil
}

func (d *DolevKnownImproved) sendInitialMessage(uid uint32, payload Size, partial bool, origin uint64) error {
	m := DolevKnownImprovedMessage{
		Src:     d.cfg.Id,
		Id:      d.cnt,
		Payload: payload,
	}

	to := d.broadcast
	if partial {
		to = d.bdPlan[origin]
	}

	for dst, p := range to {
		dp := make([]algo.DolevPath, len(p))

		for i, d := range p {
			dp[i] = algo.DolevPath{Desired: d.P, Prio: d.Prio}
		}

		m.Paths = dp
		d.n.Send(0, dst, uid, m, BroadcastInfo{})
	}

	return nil
}

func (d *DolevKnownImproved) hasDelivered(id dolevIdentifier) bool {
	_, ok := d.delivered[id]
	return ok
}

func (d *DolevKnownImproved) Receive(_ uint8, src uint64, uid uint32, data Size) {
	if d.cfg.Byz {
		// TODO: better byzantine behaviour?
		return
	}

	m := data.(DolevKnownImprovedMessage)
	id := dolevIdentifier{
		Src:        m.Src,
		Id:         m.Id,
		TrackingId: uid,
		Hash:       MustHash(m.Payload),
	}

	// Add latest edge to path for message
	for i, p := range m.Paths {
		p.Actual = append(p.Actual, simple.WeightedEdge{
			F: simple.Node(src),
			T: simple.Node(d.cfg.Id),
		})
		m.Paths[i] = p

		// Add paths to mem for this message
		if !d.hasDelivered(id) {
			d.paths[id] = append(d.paths[id], p.Actual)
		}
	}

	if d.cfg.Id == m.Src {
		panic("received message from self, should have been delivered already")
	}

	if !d.hasDelivered(id) {
		// Additional modification (based on bonomi 7): Accept messages from origin immediately
		if m.Src == src || graphs.VerifyDisjointPaths(d.paths[id], simple.Node(m.Src), simple.Node(d.cfg.Id), d.cfg.F+1) {
			//fmt.Printf("proc %v is delivering %v at %v\n", d.cfg.Id, id, time.Now())
			d.delivered[id] = struct{}{}
			d.app.Deliver(uid, m.Payload, m.Src)

			// Memory cleanup
			d.paths[id] = nil
		}
	}

	// Send to next hops
	if err := d.sendMergedMessage(uid, id, m); err != nil {
		fmt.Printf("process %v errored while sending dolev (known, improved) message: %v\n", d.cfg.Id, err)
		os.Exit(1)
	}
}

func (d *DolevKnownImproved) Broadcast(uid uint32, payload Size, bc BroadcastInfo) {
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

		partial := false
		partialId := d.cfg.Id

		if d.bd {
			partial = bc.Type == BrachaPartial
			m := payload.(brachaWrapper).msg.(BrachaMessage)
			partialId = m.Src
		}

		if err := d.sendInitialMessage(uid, payload, partial, partialId); err != nil {
			fmt.Printf("process %v errored while broadcasting dolev (known, improved) message: %v\n", d.cfg.Id, err)
			os.Exit(1)
		}

		d.cnt += 1
	}
}

func (d *DolevKnownImproved) Category() ProtocolCategory {
	return DolevCat
}
