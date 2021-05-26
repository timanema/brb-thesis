package brb

import (
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"os"
	"rp-runner/brb/algo"
	"rp-runner/graphs"
	"strconv"
)

// Dolev with routing and additional optimizations for RP Tim Anema.
// Used for testing cross-layer Bracha-Dolev optimizations without making too many breaking changes to normal Dolev.
type DolevKnownImprovedBD struct {
	n   Network
	app Application
	cfg Config

	cnt uint32

	delivered map[dolevIdentifier]struct{}
	paths     map[dolevIdentifier][]graphs.Path

	buffer map[dolevIdentifier][]algo.DolevPath

	broadcast algo.BroadcastPlan

	bd       bool
	bdPlan   map[uint64]algo.BroadcastPlan
	bdBuffer map[brachaIdentifier][]temp
}

type temp struct {
	Id   dolevIdentifier
	Type uint8
}

var _ Protocol = (*DolevKnownImprovedBD)(nil)

func (d *DolevKnownImprovedBD) Init(n Network, app Application, cfg Config) {
	d.n = n
	d.app = app
	d.cfg = cfg
	d.delivered = make(map[dolevIdentifier]struct{})
	d.paths = make(map[dolevIdentifier][]graphs.Path)
	d.buffer = make(map[dolevIdentifier][]algo.DolevPath)
	d.bdBuffer = make(map[brachaIdentifier][]temp)

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

		// TODO: remove testing info
		deps := algo.FindDependants(routes)
		//fmt.Println(deps)

		algo.FixDeadlocks(routes, deps)

		if d.bd {
			n, m := graphs.Nodes(d.cfg.Graph)
			d.bdPlan = algo.BrachaDolevRouting(routes, graphs.FindAdjMap(graphs.Directed(d.cfg.Graph), m), n, d.cfg.N, d.cfg.F)
			//fmt.Printf("proc %v: %v\n", d.cfg.Id, d.bdPlan)
		}

		d.broadcast = algo.DolevRouting(routes, true, true)
	}
}

func (d *DolevKnownImprovedBD) sendMergedMessage(uid uint32, m BrachaDolevWrapper) error {
	bdm := BrachaDolevWrapper{
		OriginalSrc:     m.OriginalSrc,
		OriginalId:      m.OriginalId,
		OriginalPayload: m.OriginalPayload,
		Included:        m.Included,
	}
	hopping := false
	bid := brachaIdentifier{
		Src:        m.OriginalSrc,
		Id:         m.OriginalId,
		TrackingId: uid,
		Hash:       MustHash(m.OriginalPayload),
	}

	for _, dm := range m.Unpack() {
		// TODO: find sensible capacity
		paths := make([]algo.DolevPath, 0, 10)

		id := dolevIdentifier{
			Src:        dm.Src,
			Id:         dm.Id,
			TrackingId: uid,
			Hash:       MustHash(dm.Payload),
		}

		del := d.hasDelivered(id)
		bw := dm.Payload.(brachaWrapper)

		// If delivered, relay all messages, including ones in the buffer
		if del || hopping {
			hopping = true
			paths = append(paths, dm.Paths...)
			paths = append(paths, d.buffer[id]...)

			// Clear buffer
			d.buffer[id] = nil
		} else {
			// If not delivered, add all messages with no priority to the buffer and
			// all messages with priority to the outgoing paths
			for _, p := range dm.Paths {
				if len(p.Desired) == len(p.Actual) {
					continue
				}

				if p.Prio {
					paths = append(paths, p)
				} else {
					d.buffer[id] = append(d.buffer[id], p)
					d.bdBuffer[bid] = append(d.bdBuffer[bid], temp{
						Id:   id,
						Type: bw.messageType,
					})
				}
			}
		}

		bm := BrachaDolevMessage{
			Src:   dm.Src,
			Id:    dm.Id,
			Type:  bw.messageType,
			Paths: paths,
		}

		if len(bm.Paths) > 0 {
			bdm.Msgs = append(bdm.Msgs, bm)
		}
	}

	if hopping {
		for _, id := range d.bdBuffer[bid] {
			bm := BrachaDolevMessage{
				Src:   id.Id.Src,
				Id:    id.Id.Id,
				Type:  id.Type,
				Paths: d.buffer[id.Id],
			}

			d.buffer[id.Id] = nil
			if len(bm.Paths) > 0 {
				bdm.Msgs = append(bdm.Msgs, bm)
			}
		}
	}

	// TODO: let paths with the same next hop join if a message is traveling there
	next := d.prepareBrachaDolevMergedPaths(bdm)

	if len(next) > 0 {
		d.n.TriggerStat(uid, StartRelay)
	}

	for dst, m := range next {
		d.n.Send(0, dst, uid, m, BroadcastInfo{})
	}

	return nil
}

func (d *DolevKnownImprovedBD) prepareBrachaDolevMergedPaths(bdm BrachaDolevWrapper) map[uint64]DolevKnownImprovedMessage {
	paths := make(map[uint64][]algo.DolevPath)
	bds := make(map[uint64][]BrachaDolevMessage)

	for _, bm := range bdm.Msgs {
		res := make(map[uint64][]algo.DolevPath)
		for _, p := range bm.Paths {
			if cur := len(p.Actual); len(p.Desired) > cur {
				// Make a copy of the path
				cp := algo.DolevPath{
					Desired: make(graphs.Path, len(p.Desired)),
					Actual:  make(graphs.Path, len(p.Actual)),
					Prio:    p.Prio,
				}

				copy(cp.Desired, p.Desired)
				copy(cp.Actual, p.Actual)

				next := uint64(p.Desired[cur].To().ID())
				paths[next] = append(paths[next], cp)
				res[next] = append(res[next], cp)
			}
		}

		for dst, paths := range res {
			bds[dst] = append(bds[dst], BrachaDolevMessage{
				Src:   bm.Src,
				Id:    bm.Id,
				Type:  bm.Type,
				Paths: paths,
			})
		}
	}

	res := make(map[uint64]DolevKnownImprovedMessage)

	for dst, msgs := range bds {
		res[dst] = DolevKnownImprovedMessage{
			Src: d.cfg.Id,
			Id:  0,
			Payload: BrachaDolevWrapper{
				Msgs:            msgs,
				OriginalSrc:     bdm.OriginalSrc,
				OriginalId:      bdm.OriginalId,
				OriginalPayload: bdm.OriginalPayload,
				Included:        bdm.Included,
			},
			Paths: paths[dst],
		}
	}

	return res
}

func (d *DolevKnownImprovedBD) sendInitialMessage(uid uint32, payload interface{}, partial bool, origin uint64) error {
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

func (d *DolevKnownImprovedBD) hasDelivered(id dolevIdentifier) bool {
	_, ok := d.delivered[id]
	return ok
}

func (d *DolevKnownImprovedBD) Receive(_ uint8, src uint64, uid uint32, data interface{}) {
	if d.cfg.Byz {
		// TODO: better byzantine behaviour?
		return
	}

	dm := data.(DolevKnownImprovedMessage)
	bdw, ok := dm.Payload.(BrachaDolevWrapper)

	msgs := []DolevKnownImprovedMessage{dm}
	if ok {
		msgs = bdw.Unpack()

		if len(msgs) > 1 {
			fmt.Println("Bracha-Dolev messages were actually merged!")
		}
	}

	for _, m := range msgs {
		id := dolevIdentifier{
			Src:        m.Src,
			Id:         m.Id,
			TrackingId: uid,
			Hash:       MustHash(m.Payload),
		}

		// Add latest edge to path for message
		for i, p := range m.Paths {
			// TODO: clean up this crazyness
			if len(p.Actual) > len(p.Desired) {
				panic("actual path is longer than desired path!")
			}

			eq := len(p.Actual) == len(p.Desired)
			check := len(p.Actual)
			if eq {
				check = len(p.Desired) - 1
			}

			if p.Desired[check].To().ID() != int64(d.cfg.Id) {
				continue
			}

			if eq {
				panic("received incorrect path!")
			}

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
	}

	// Send to next hops
	bw := Pack(msgs)

	if err := d.sendMergedMessage(uid, bw); err != nil {
		fmt.Printf("process %v errored while sending dolev (known, improved) message: %v\n", d.cfg.Id, err)
		os.Exit(1)
	}
}

func (d *DolevKnownImprovedBD) Broadcast(uid uint32, payload interface{}, bc BroadcastInfo) {
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
			m := payload.(brachaWrapper).msg.(BrachaImprovedMessage)
			partialId = m.Src
		}

		if err := d.sendInitialMessage(uid, payload, partial, partialId); err != nil {
			fmt.Printf("process %v errored while broadcasting dolev (known, improved) message: %v\n", d.cfg.Id, err)
			os.Exit(1)
		}

		d.cnt += 1
	}
}

func (d *DolevKnownImprovedBD) Category() ProtocolCategory {
	return DolevCat
}
