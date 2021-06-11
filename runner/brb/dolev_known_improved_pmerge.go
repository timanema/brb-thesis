package brb

import (
	"crypto/sha256"
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"rp-runner/brb/algo"
	"rp-runner/graphs"
	"strconv"
)

// Dolev with routing and additional optimizations for RP Tim Anema.
type DolevKnownImprovedPM struct {
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
	bdBuffer map[brachaIdentifier][]bdBufferEntry

	similarPayloads map[[sha256.Size]byte]map[dolevIdentifier]struct{}
}

var _ Protocol = (*DolevKnownImprovedPM)(nil)

func (d *DolevKnownImprovedPM) Init(n Network, app Application, cfg Config) {
	d.n = n
	d.app = app
	d.cfg = cfg
	d.delivered = make(map[dolevIdentifier]struct{})
	d.paths = make(map[dolevIdentifier][]graphs.Path)
	d.buffer = make(map[dolevIdentifier][]algo.DolevPath)
	d.bdBuffer = make(map[brachaIdentifier][]bdBufferEntry)
	d.similarPayloads = make(map[[sha256.Size]byte]map[dolevIdentifier]struct{})

	if !cfg.Silent && cfg.Byz {
		fmt.Printf("process %v is a Dolev (known improved) Byzantine node\n", cfg.Id)
		return
	}

	if d.broadcast == nil && !d.cfg.Unused {
		_, bd := cfg.AdditionalConfig.(BrachaDolevConfig)
		d.bd = bd

		w := 0
		if d.cfg.OptimizationConfig.DolevReusePaths {
			w = d.cfg.N / 10
		}

		routes, err := algo.BuildRoutingTable(cfg.Graph, graphs.Node{
			Id:   int64(d.cfg.Id),
			Name: strconv.Itoa(int(d.cfg.Id)),
		}, d.cfg.F*2+1, w, d.cfg.OptimizationConfig.DolevSingleHopNeighbour)
		if err != nil {
			panic(fmt.Sprintf("process %v errored while building lookup table: %v\n", d.cfg.Id, err))
		}

		algo.FixDeadlocks(routes)

		if d.bd {
			n, m := graphs.Nodes(d.cfg.Graph)
			d.bdPlan = algo.BrachaDolevRouting(routes, graphs.FindAdjMap(graphs.Directed(d.cfg.Graph), m), n, d.cfg.N, d.cfg.F)
		}

		d.broadcast = algo.DolevRouting(routes, d.cfg.OptimizationConfig.DolevCombineNextHops, d.cfg.OptimizationConfig.DolevFilterSubpaths)
	}
}

func (d *DolevKnownImprovedPM) sendMergedBDMessage(uid uint32, m BrachaDolevWrapperMsg) {
	bdm := BrachaDolevWrapperMsg{
		OriginalSrc:     m.OriginalSrc,
		OriginalId:      m.OriginalId,
		OriginalPayload: m.OriginalPayload,
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
					d.bdBuffer[bid] = append(d.bdBuffer[bid], bdBufferEntry{
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

	next := d.prepareBrachaDolevMergedPaths(bdm)

	if len(next) > 0 {
		d.n.TriggerStat(uid, StartRelay)
	}

	for dst, m := range next {
		d.n.Send(0, dst, uid, m, BroadcastInfo{})
	}
}

func (d *DolevKnownImprovedPM) prepareBrachaDolevMergedPaths(bdm BrachaDolevWrapperMsg) map[uint64]DolevKnownImprovedMessage {
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
			Payload: BrachaDolevWrapperMsg{
				Msgs:            msgs,
				OriginalSrc:     bdm.OriginalSrc,
				OriginalId:      bdm.OriginalId,
				OriginalPayload: bdm.OriginalPayload,
			},
		}
	}

	return res
}

type hopInfo struct {
	next algo.NextHopPlan
	id   dolevIdentifier
}

func (d *DolevKnownImprovedPM) getSimilarKnown(next algo.NextHopPlan, id dolevIdentifier) []hopInfo {
	similar, ok := d.similarPayloads[id.Hash]
	if !ok {
		return nil
	}

	var res []hopInfo

	for sId := range similar {
		n := make(algo.NextHopPlan)
		buf, _ := algo.SetPiggybacks(next, n, d.buffer[sId])
		d.buffer[sId] = buf

		for dst, paths := range n {
			res := make([]algo.DolevPath, 0, len(paths))
			for _, p := range paths {
				// Make a copy of the path
				cp := algo.DolevPath{
					Desired: make(graphs.Path, len(p.Desired)),
					Actual:  make(graphs.Path, len(p.Actual)),
					Prio:    p.Prio,
				}

				copy(cp.Desired, p.Desired)
				copy(cp.Actual, p.Actual)
				res = append(res, cp)
			}
			n[dst] = res
		}

		if len(n) > 0 {
			res = append(res, hopInfo{
				next: n,
				id:   sId,
			})
		}
	}

	return res
}

func (d *DolevKnownImprovedPM) sendMergedPayload(uid uint32, bufferCnt map[uint64]int,
	next algo.NextHopPlan, m DolevKnownImprovedMessage, hoppers []hopInfo) {
	payload := m.Payload

	for dst, p := range next {
		msgs := make([]dolevWrapperWrapper, 0, len(hoppers))

		for _, hopper := range hoppers {
			for hDst, p := range hopper.next {
				if dst == hDst {
					msgs = append(msgs, dolevWrapperWrapper{
						Src:        hopper.id.Src,
						Id:         hopper.id.Id,
						TrackingId: hopper.id.TrackingId,
						Paths:      p,
					})
				}
			}
		}

		m.Paths = p

		if len(msgs) > 0 {
			wrapper := DolevWrapperMessage{
				Msgs:    msgs,
				Payload: payload,
			}
			m.Payload = wrapper
		} else {
			m.Payload = payload
		}

		// For statistics purposes, check if the buffer items were actually merged
		if len(p) > bufferCnt[dst] {
			for i := 0; i < bufferCnt[dst]; i++ {
				d.n.TriggerStat(uid, DolevPathMerge)
			}
		} else if len(p) == bufferCnt[dst] {
			for i := 1; i < bufferCnt[dst]; i++ {
				d.n.TriggerStat(uid, DolevPathMerge)
			}
		}

		d.n.Send(0, dst, uid, m, BroadcastInfo{})
	}
}

func (d *DolevKnownImprovedPM) sendMergedMessage(uid uint32, m DolevKnownImprovedMessage) {
	id := dolevIdentifier{
		Src:        m.Src,
		Id:         m.Id,
		TrackingId: uid,
		Hash:       MustHash(m.Payload),
	}
	del := d.hasDelivered(id)

	paths := make([]algo.DolevPath, 0, len(m.Paths))
	bufferCnt := make(map[uint64]int)

	// If delivered, relay all messages, including ones in the buffer
	if del || !d.cfg.OptimizationConfig.DolevRelayMerging {
		paths = append(paths, m.Paths...)

		if buf, ok := d.buffer[id]; ok {
			paths = append(paths, buf...)

			// For statistics purposes
			for _, p := range buf {
				if cur := len(p.Actual); len(p.Desired) > cur {
					next := uint64(p.Desired[cur].To().ID())
					bufferCnt[next] += 1
				}
			}
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

	next := algo.CombineDolevPaths(paths)
	if len(next) == 0 {
		return
	}

	if d.cfg.OptimizationConfig.DolevRelayMerging {
		buf, piggyBacks := algo.AddPiggybacks(next, d.buffer[id])
		d.buffer[id] = buf

		for i := 0; i < piggyBacks; i++ {
			d.n.TriggerStat(uid, DolevPathMerge)
		}
	}

	if len(next) > 0 {
		d.n.TriggerStat(uid, StartRelay)
	}

	if d.cfg.OptimizationConfig.DolevPayloadMerging {
		if r := d.getSimilarKnown(next, id); len(r) > 0 {
			d.sendMergedPayload(uid, bufferCnt, next, m, r)

			return
		}
	}

	for dst, p := range next {
		// For statistics purposes, check if the buffer items were actually merged
		if len(p) > bufferCnt[dst] {
			for i := 0; i < bufferCnt[dst]; i++ {
				d.n.TriggerStat(uid, DolevPathMerge)
			}
		} else if len(p) == bufferCnt[dst] {
			for i := 1; i < bufferCnt[dst]; i++ {
				d.n.TriggerStat(uid, DolevPathMerge)
			}
		}

		m.Paths = p
		d.n.Send(0, dst, uid, m, BroadcastInfo{})
	}
}

func (d *DolevKnownImprovedPM) sendInitialMessage(uid uint32, payload Size, partial bool, origin uint64) {
	m := DolevKnownImprovedMessage{
		Src:     d.cfg.Id,
		Id:      d.cnt,
		Payload: payload,
	}

	to := d.broadcast
	if partial && d.cfg.OptimizationConfig.BrachaDolevPartialBroadcast {
		to = d.bdPlan[origin]
	}

	for dst, p := range to {
		if d.cfg.OptimizationConfig.DolevCombineNextHops {
			dp := make([]algo.DolevPath, len(p))

			for i, d := range p {
				dp[i] = algo.DolevPath{Desired: d.P, Prio: d.Prio}
			}

			m.Paths = dp
			d.n.Send(0, dst, uid, m, BroadcastInfo{})
		} else {
			for _, path := range p {
				m.Paths = []algo.DolevPath{{Desired: path.P, Prio: path.Prio}}

				d.n.Send(0, dst, uid, m, BroadcastInfo{})
			}
		}
	}
}

func (d *DolevKnownImprovedPM) hasDelivered(id dolevIdentifier) bool {
	_, ok := d.delivered[id]
	return ok
}

func (d *DolevKnownImprovedPM) checkPayloadSimilarity(id dolevIdentifier) {
	if _, ok := d.similarPayloads[id.Hash]; !ok {
		d.similarPayloads[id.Hash] = make(map[dolevIdentifier]struct{})
	}

	d.similarPayloads[id.Hash][id] = struct{}{}
}

func (d *DolevKnownImprovedPM) Receive(_ uint8, src uint64, uid uint32, data Size) {
	if d.cfg.Byz {
		// TODO: better byzantine behaviour?
		return
	}

	dm := data.(DolevKnownImprovedMessage)
	bdw, bdWrapperOk := dm.Payload.(BrachaDolevWrapperMsg)
	dpw, dpWrapperOk := dm.Payload.(DolevWrapperMessage)
	msgs := []DolevKnownImprovedMessage{dm}
	var tracking []uint32

	if bdWrapperOk {
		msgs = bdw.Unpack()

		for i := 1; i < len(msgs); i++ {
			d.n.TriggerStat(uid, BrachaDolevMerge)
		}
	} else if dpWrapperOk {
		msgs, tracking = dpw.Unpack(dm, uid)

		for i := 1; i < len(msgs); i++ {
			d.n.TriggerStat(uid, DolevPayloadMerge)
		}
	}

	for i, m := range msgs {
		track := uid
		if len(tracking) > 0 {
			track = tracking[i]
		}

		id := dolevIdentifier{
			Src:        m.Src,
			Id:         m.Id,
			TrackingId: track,
			Hash:       MustHash(m.Payload),
		}

		if d.cfg.OptimizationConfig.DolevPayloadMerging {
			d.checkPayloadSimilarity(id)
		}

		// Add latest edge to path for message
		for i, p := range m.Paths {
			if len(p.Actual) >= len(p.Desired) {
				panic("actual path is longer than desired path!")
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
				d.app.Deliver(track, m.Payload, m.Src)

				// Memory cleanup
				d.paths[id] = nil
			}
		}
	}

	switch {
	case dpWrapperOk:
		// TODO: think of something better for this
		for i, m := range msgs {
			track := uid
			if len(tracking) > 0 {
				track = tracking[i]
			}

			// Send to next hops
			d.sendMergedMessage(track, m)
		}
	case d.bd:
		// Send to next hops
		bw := Pack(msgs)

		if d.cfg.OptimizationConfig.BrachaDolevMerge {
			d.sendMergedBDMessage(uid, bw)
		} else {
			for _, m := range msgs {
				// Send to next hops
				d.sendMergedMessage(uid, m)
			}
		}
	default:
		// Send to next hops
		d.sendMergedMessage(uid, dm)
	}
}

func (d *DolevKnownImprovedPM) Broadcast(uid uint32, payload Size, bc BroadcastInfo) {
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

		//fmt.Printf("%v is sending %v (%v)\n", d.cfg.Id, uid, d.cnt)
		d.sendInitialMessage(uid, payload, partial, partialId)

		d.cnt += 1
	}
}

func (d *DolevKnownImprovedPM) Category() ProtocolCategory {
	return DolevCat
}
