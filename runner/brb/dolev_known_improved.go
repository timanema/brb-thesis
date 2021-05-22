package brb

import (
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"os"
	"rp-runner/graphs"
	"strconv"
)

type DolevKnownImprovedMessage struct {
	Src     uint64
	Id      uint32
	Payload interface{}
	Paths   []dolevPath
}

// Dolev with routing and additional optimizations for RP Tim Anema
type DolevKnownImproved struct {
	n   Network
	app Application
	cfg Config

	cnt uint32

	delivered map[dolevIdentifier]struct{}
	paths     map[dolevIdentifier][]graphs.Path
	//routes    map[uint64][]graphs.Path
	broadcast map[uint64][]graphs.Path
}

var _ Protocol = (*DolevKnownImproved)(nil)

func (d *DolevKnownImproved) Init(n Network, app Application, cfg Config) {
	d.n = n
	d.app = app
	d.cfg = cfg
	d.delivered = make(map[dolevIdentifier]struct{})
	d.paths = make(map[dolevIdentifier][]graphs.Path)

	if !cfg.Silent && cfg.Byz {
		fmt.Printf("process %v is a Dolev (known improved) Byzantine node\n", cfg.Id)
		return
	}

	if d.broadcast == nil && !d.cfg.Unused {
		// Additional modification (based on bonomi 7): Accept messages from origin immediately, neighbours only need one path
		routes, err := graphs.BuildLookupTable(cfg.Graph, graphs.Node{
			Id:   int64(d.cfg.Id),
			Name: strconv.Itoa(int(d.cfg.Id)),
		}, d.cfg.F*2+1, 5, true)
		if err != nil {
			fmt.Printf("process %v errored while building lookup table: %v\n", d.cfg.Id, err)
			os.Exit(1)
		}
		br := make([]graphs.Path, 0, len(routes))

		for _, g := range routes {
			br = append(br, g...)
		}

		d.broadcast = combinePaths(graphs.FilterSubpaths(br))
		// TODO: check if there can be a case where a node is in multiple next-hop chains (if so, merging them later on can improve perf).
	}

	fmt.Printf("")
}

func combinePaths(paths []graphs.Path) map[uint64][]graphs.Path {
	res := make(map[uint64][]graphs.Path)

	for _, p := range paths {
		next := uint64(p[0].To().ID())
		res[next] = append(res[next], p)
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

func (d *DolevKnownImproved) sendMergedMessage(uid uint32, m DolevKnownImprovedMessage) error {
	next := combineDolevPaths(m.Paths)

	if len(next) > 0 {
		d.n.TriggerStat(uid, StartRelay)
	}

	for dst, p := range next {
		m.Paths = p
		d.n.Send(0, dst, uid, m, BroadcastInfo{})
	}

	return nil
}

func (d *DolevKnownImproved) sendInitialMessage(uid uint32, payload interface{}) error {
	m := DolevKnownImprovedMessage{
		Src:     d.cfg.Id,
		Id:      d.cnt,
		Payload: payload,
	}

	for dst, p := range d.broadcast {
		dp := make([]dolevPath, len(p))

		for i, d := range p {
			dp[i] = dolevPath{Desired: d}
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

func (d *DolevKnownImproved) Receive(_ uint8, src uint64, uid uint32, data interface{}) {
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

	// Send to next hops
	if err := d.sendMergedMessage(uid, m); err != nil {
		fmt.Printf("process %v errored while sending dolev (known, improved) message: %v\n", d.cfg.Id, err)
		os.Exit(1)
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
			delete(d.paths, id)
		}
	}
}

func (d *DolevKnownImproved) Broadcast(uid uint32, payload interface{}) {
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
			fmt.Printf("process %v errored while broadcasting dolev (known, improved) message: %v\n", d.cfg.Id, err)
			os.Exit(1)
		}

		d.cnt += 1
	}
}

func (d *DolevKnownImproved) Category() ProtocolCategory {
	return DolevCat
}

/*

2:[[{5 2} {2 1}] [{5 2} {2 4}] [{5 2} {2 0} {0 9}] [{5 2} {2 0} {0 8}]]
3:[[{5 3} {3 1}] [{5 3} {3 4}] [{5 3} {3 0}]]
6:[[{5 6} {6 8} {8 1}] [{5 6} {6 4}] [{5 6} {6 9}] [{5 6} {6 8} {8 0}]]
7:[[{5 7} {7 9}] [{5 7} {7 8}]]


*/
