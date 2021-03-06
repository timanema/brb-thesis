package brb

import (
	"crypto/sha256"
	"fmt"
	"gonum.org/v1/gonum/graph/simple"
	"reflect"
	"rp-runner/brb/algo"
	"rp-runner/graphs"
)

type DolevKnownImprovedMessage struct {
	Src     uint64
	Id      uint32
	Payload Size
	Paths   []algo.DolevPath
	Partial bool
}

func (d DolevKnownImprovedMessage) SizeOf() uintptr {
	return reflect.TypeOf(d.Src).Size() + reflect.TypeOf(d.Id).Size() + algo.SizeOfMultiplePaths(d.Paths) + d.Payload.SizeOf()
}

type bdBufferEntry struct {
	Id      dolevIdentifier
	Type    uint8
	Partial bool
	Buf     algo.DolevPath
}

// Dolev with routing and additional optimizations for RP Tim Anema.
type DolevKnownImproved struct {
	n   Network
	app Application
	cfg Config

	cnt uint32

	delivered map[dolevIdentifier]struct{}
	paths     map[dolevIdentifier][]graphs.Path

	buffer        map[dolevIdentifier][]algo.DolevPath
	partialBuffer map[dolevIdentifier][]algo.DolevPath

	implicitPathsUsed map[dolevIdentifier][]algo.DolevPath

	broadcast algo.BroadcastPlan

	bd       bool
	bdPlan   map[uint64]algo.BroadcastPlan
	bdBuffer map[brachaIdentifier][]bdBufferEntry

	similarPayloads map[[sha256.Size]byte]map[dolevIdentifier]struct{}
}

var _ Protocol = (*DolevKnownImproved)(nil)

func (d *DolevKnownImproved) Init(n Network, app Application, cfg Config) {
	d.n = n
	d.app = app
	d.cfg = cfg
	d.delivered = make(map[dolevIdentifier]struct{})
	d.paths = make(map[dolevIdentifier][]graphs.Path)
	d.buffer = make(map[dolevIdentifier][]algo.DolevPath)
	d.partialBuffer = make(map[dolevIdentifier][]algo.DolevPath)
	d.implicitPathsUsed = make(map[dolevIdentifier][]algo.DolevPath)
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

		if d.cfg.OptimizationConfig.DolevImplicitPath {
			d.broadcast = d.cfg.Precomputed.FullTable.Plan[d.cfg.Id]
			d.bdPlan = d.cfg.Precomputed.FullTable.BDPlan[d.cfg.Id]
		} else {
			d.broadcast, d.bdPlan = algo.Routing(nil, d.cfg.Id, d.cfg.Graph, w, d.cfg.N, d.cfg.F,
				d.cfg.OptimizationConfig.DolevSingleHopNeighbour, d.cfg.OptimizationConfig.DolevCombineNextHops,
				d.cfg.OptimizationConfig.DolevFilterSubpaths, d.bd)
		}
	}
}

func (d *DolevKnownImproved) getPaths(m DolevKnownImprovedMessage, id dolevIdentifier) []algo.DolevPath {
	var partialId uint64
	var partial bool

	w, ok := m.Payload.(brachaWrapper)
	if ok {
		if msg, ok := w.msg.(BrachaMessage); d.bd && ok && m.Partial {
			partialId = msg.Src
			partial = true
		}
	}

	if d.cfg.OptimizationConfig.DolevImplicitPath {
		paths := d.cfg.Precomputed.FullTable.FindAllMatches(m.Src, partialId, m.Paths, partial, d.implicitPathsUsed[id])
		d.implicitPathsUsed[id] = append(d.implicitPathsUsed[id], paths...)
		return paths
	} else {
		return m.Paths
	}
}

func (d *DolevKnownImproved) sendMergedBDMessage(uid uint32, m BrachaDolevWrapperMsg) {
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

	// When a certain combination of optimizations is enabled, some special care needs to be taken as to not
	// delete path data (more specifically the partial path flag) when filling buffers.
	optimCombi := d.cfg.OptimizationConfig.DolevImplicitPath &&
		d.cfg.OptimizationConfig.BrachaDolevPartialBroadcast &&
		d.cfg.OptimizationConfig.BrachaDolevMerge

	for _, dm := range m.Unpack() {
		var paths []algo.DolevPath

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
			paths = append(paths, d.getPaths(dm, id)...)
			if !dm.Partial || !optimCombi {
				paths = append(paths, d.buffer[id]...)
				d.buffer[id] = nil
			} else {
				paths = append(paths, d.partialBuffer[id]...)
				d.partialBuffer[id] = nil
			}
		} else {
			// If not delivered, add all messages with no priority to the buffer and
			// all messages with priority to the outgoing paths
			for _, p := range d.getPaths(dm, id) {
				if len(p.Desired) == len(p.Actual) {
					continue
				}

				if p.Prio {
					paths = append(paths, p)
				} else {
					if !dm.Partial || !optimCombi {
						d.buffer[id] = append(d.buffer[id], p)
					} else {
						d.partialBuffer[id] = append(d.partialBuffer[id], p)
					}

					d.bdBuffer[bid] = append(d.bdBuffer[bid], bdBufferEntry{
						Id:      id,
						Type:    bw.messageType,
						Partial: dm.Partial,
					})
				}
			}
		}

		bm := BrachaDolevMessage{
			Src:     dm.Src,
			Id:      dm.Id,
			Type:    bw.messageType,
			Paths:   paths,
			Partial: dm.Partial,
		}

		if len(bm.Paths) > 0 {
			bdm.Msgs = append(bdm.Msgs, bm)
		}
	}

	if hopping {
		for _, id := range d.bdBuffer[bid] {
			if id.Partial && optimCombi {
				bmPartial := BrachaDolevMessage{
					Src:     id.Id.Src,
					Id:      id.Id.Id,
					Type:    id.Type,
					Paths:   d.partialBuffer[id.Id],
					Partial: id.Partial,
				}
				d.partialBuffer[id.Id] = nil

				if len(bmPartial.Paths) > 0 {
					bdm.Msgs = append(bdm.Msgs, bmPartial)
				}
			}

			bm := BrachaDolevMessage{
				Src:     id.Id.Src,
				Id:      id.Id.Id,
				Type:    id.Type,
				Paths:   d.buffer[id.Id],
				Partial: id.Partial && !(id.Partial && optimCombi),
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
		d.Send(uid, dst, m)
	}
}

func (d *DolevKnownImproved) prepareBrachaDolevMergedPaths(bdm BrachaDolevWrapperMsg) map[uint64]DolevKnownImprovedMessage {
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
			if d.cfg.OptimizationConfig.DolevImplicitPath {
				paths = algo.CleanDesiredPaths(paths)
			}

			bds[dst] = append(bds[dst], BrachaDolevMessage{
				Src:     bm.Src,
				Id:      bm.Id,
				Type:    bm.Type,
				Paths:   paths,
				Partial: bm.Partial,
			})
		}
	}

	res := make(map[uint64]DolevKnownImprovedMessage)

	for dst, msgs := range bds {
		res[dst] = DolevKnownImprovedMessage{
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

func (d *DolevKnownImproved) getSimilarKnown(next algo.NextHopPlan, id dolevIdentifier) []hopInfo {
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

func (d *DolevKnownImproved) sendMergedPayload(uid uint32, bufferCnt map[uint64]int,
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

		d.Send(uid, dst, m)
	}
}

func (d *DolevKnownImproved) sendMergedMessage(uid uint32, m DolevKnownImprovedMessage) {
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
		paths = append(paths, d.getPaths(m, id)...)

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
		for _, p := range d.getPaths(m, id) {
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
		d.Send(uid, dst, m)
	}
}

func (d *DolevKnownImproved) sendInitialMessage(uid uint32, payload Size, partial bool, origin uint64) {
	m := DolevKnownImprovedMessage{
		Src:     d.cfg.Id,
		Id:      d.cnt,
		Payload: payload,
	}

	to := d.broadcast
	if partial && d.cfg.OptimizationConfig.BrachaDolevPartialBroadcast {
		to = d.bdPlan[origin]
		m.Partial = true
	}

	for dst, p := range to {
		if d.cfg.OptimizationConfig.DolevCombineNextHops {
			dp := make([]algo.DolevPath, len(p))

			for i, dol := range p {
				dp[i] = algo.DolevPath{Desired: dol.P, Prio: dol.Prio}
			}

			m.Paths = dp
			d.Send(uid, dst, m)
		} else {
			for _, path := range p {
				m.Paths = []algo.DolevPath{{Desired: path.P, Prio: path.Prio}}
				d.Send(uid, dst, m)
			}
		}
	}
}

func (d *DolevKnownImproved) Send(uid uint32, dst uint64, m DolevKnownImprovedMessage) {
	if d.cfg.OptimizationConfig.DolevImplicitPath {
		m.Paths = algo.CleanDesiredPaths(m.Paths)
	}

	d.n.Send(0, dst, uid, m, BroadcastInfo{})
}

func (d *DolevKnownImproved) hasDelivered(id dolevIdentifier) bool {
	_, ok := d.delivered[id]
	return ok
}

func (d *DolevKnownImproved) checkPayloadSimilarity(id dolevIdentifier) {
	if _, ok := d.similarPayloads[id.Hash]; !ok {
		d.similarPayloads[id.Hash] = make(map[dolevIdentifier]struct{})
	}

	d.similarPayloads[id.Hash][id] = struct{}{}
}

func (d *DolevKnownImproved) Receive(_ uint8, src uint64, uid uint32, data Size) {
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

		//fmt.Printf("%v is sending %v (%v)\n", d.cfg.Id, uid, d.cnt)
		d.sendInitialMessage(uid, payload, partial, partialId)

		d.cnt += 1
	}
}

func (d *DolevKnownImproved) Category() ProtocolCategory {
	return DolevCat
}
