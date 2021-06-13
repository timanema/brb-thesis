package ctrl

import (
	"fmt"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"math/rand"
	"os"
	"reflect"
	"rp-runner/brb"
	"rp-runner/brb/algo"
	"rp-runner/msg"
	"rp-runner/process"
	"sync"
	"time"
)

type proc struct {
	p                 *process.Process
	alive, ready, byz bool

	err error
}

type Config struct {
	CtrlBuffer, ProcBuffer int
	PollDelay              time.Duration
}

type Controller struct {
	ctl      chan process.Message
	channels map[uint64]chan process.Message

	cfg Config

	stopCh chan struct{}

	p     map[uint64]proc
	pLock sync.Mutex

	payloadMap map[uint32]interface{}
	deliverMap map[uint32]map[uint64]struct{}
	sendMap    map[uint32]time.Time
	dLock      sync.Mutex

	al, rdy int
}

func StartController(cfg Config) (*Controller, error) {
	c := &Controller{
		ctl:        make(chan process.Message, cfg.CtrlBuffer),
		channels:   make(map[uint64]chan process.Message),
		cfg:        cfg,
		stopCh:     make(chan struct{}),
		p:          make(map[uint64]proc),
		payloadMap: make(map[uint32]interface{}),
		deliverMap: make(map[uint32]map[uint64]struct{}),
		sendMap:    make(map[uint32]time.Time),
	}
	go c.run()

	return c, nil
}

func (c *Controller) startProcess(cfg process.Config, bp brb.Protocol) error {
	c.pLock.Lock()
	p, err := process.StartProcess(cfg.ByzConfig.Id, cfg, c.stopCh, cfg.ByzConfig.Neighbours, bp, c.ctl)
	if err != nil {
		return errors.Wrap(err, "unable to start process")
	}

	c.p[cfg.ByzConfig.Id] = proc{p: p, byz: cfg.ByzConfig.Byz}
	c.channels[cfg.ByzConfig.Id] = make(chan process.Message, c.cfg.ProcBuffer)
	c.pLock.Unlock()
	return nil
}

func (c *Controller) FlushProcesses() {
	c.pLock.Lock()
	defer c.pLock.Unlock()

	for _, p := range c.p {
		p.p.Flush()
	}

	// TODO: make better
	time.Sleep(time.Second * 3)

	for _, p := range c.p {
		p.p.StopFlush()
	}

	// To ensure all flushing routines are stopped
	time.Sleep(time.Millisecond * 400)
}

// TODO: random byzantine nodes?
func (c *Controller) StartProcesses(cfg process.Config, opt brb.OptimizationConfig, g *simple.WeightedUndirectedGraph, bp brb.Protocol, F int, possibleTransmitters []uint64, allTransmit bool) error {
	nodes := g.Nodes()
	byzLeft := F
	N := nodes.Len()

	transmitCheck := make(map[uint64]struct{}, len(possibleTransmitters))
	for _, t := range possibleTransmitters {
		transmitCheck[t] = struct{}{}
	}

	if r := len(transmitCheck); N-r < F {
		return errors.Errorf("not enough nodes to support %v possible transmitters with %v byzantine nodes", r, F)
	}

	var fullTable *algo.FullRoutingTable
	if opt.DolevImplicitPath {
		w := 0
		if opt.DolevReusePaths {
			w = N / 10
		}

		var err error
		fullTable, err = algo.BuildFullRoutingTable(g, w, N, F, F*2+1, opt.DolevSingleHopNeighbour,
			opt.DolevCombineNextHops, opt.DolevFilterSubpaths, bp.Category() == brb.BrachaDolevCat)
		if err != nil {
			return errors.Wrap(err, "failed to build full routing table")
		}
	}

	for nodes.Next() {
		n := nodes.Node()
		to := g.From(n.ID())
		neighbours := make([]uint64, 0, to.Len())

		for to.Next() {
			neighbours = append(neighbours, uint64(to.Node().ID()))
		}

		pg := simple.NewWeightedUndirectedGraph(0, 0)
		graph.CopyWeighted(pg, g)

		_, possibleTransmitter := transmitCheck[uint64(n.ID())]
		byz := byzLeft > 0 && !possibleTransmitter
		if byz {
			byzLeft -= 1
		}

		pcfg := cfg
		pcfg.ByzConfig = brb.Config{
			Byz:                byz,
			F:                  F,
			N:                  N,
			Id:                 uint64(n.ID()),
			Neighbours:         neighbours,
			Graph:              pg,
			Unused:             !possibleTransmitter,
			OptimizationConfig: opt,
			Precomputed:        brb.PrecomputedValues{FullTable: fullTable},
		}

		if allTransmit {
			pcfg.ByzConfig.Unused = false
		}

		if err := c.startProcess(pcfg, reflect.New(reflect.ValueOf(bp).Elem().Type()).Interface().(brb.Protocol)); err != nil {
			return errors.Wrap(err, "failed to create process")
		}
	}

	c.pLock.Lock()
	for _, p := range c.p {
		if err := p.p.Start(c.channels); err != nil {
			return errors.Wrap(err, "failed to start process")
		}
	}
	c.pLock.Unlock()

	return nil
}

func (c *Controller) TriggerMessageSend(id uint64, payload brb.Size) (uint32, error) {
	uid := rand.Uint32()

	m := msg.TriggerMessage{Id: uid, Payload: payload}

	c.pLock.Lock()
	if _, ok := c.channels[id]; !ok {
		c.pLock.Unlock()
		return 0, errors.New("invalid origin node")
	}
	c.pLock.Unlock()

	c.dLock.Lock()
	c.payloadMap[uid] = payload
	c.deliverMap[uid] = make(map[uint64]struct{})
	c.sendMap[uid] = time.Now()
	c.dLock.Unlock()

	c.send(id, msg.TriggerMessageType, m)

	return uid, nil
}

func (c *Controller) WaitForAlive() error {
	waiting := true
	for waiting {
		waiting = false
		c.pLock.Lock()
		for pic, p := range c.p {
			if p.err != nil {
				return errors.Wrapf(p.err, "process %v failed", pic)
			} else if !p.alive {
				fmt.Printf("waiting for %v alive\n", pic)
				waiting = true
				time.Sleep(c.cfg.PollDelay)
				break
			}
		}
		c.pLock.Unlock()
	}

	return nil
}

func (c *Controller) WaitForReady() error {
	waiting := true
	for waiting {
		waiting = false
		c.pLock.Lock()
		for pic, p := range c.p {
			if p.err != nil {
				return errors.Wrapf(p.err, "process %v failed", pic)
			} else if !p.ready {
				//fmt.Printf("waiting for %v ready\n", pic)
				waiting = true
				time.Sleep(c.cfg.PollDelay)
				break
			}
		}
		c.pLock.Unlock()
	}

	return nil
}

func (c *Controller) WaitForDeliver(uid uint32) Stats {
	c.pLock.Lock()
	needed := make(map[uint64]struct{})
	for pid, p := range c.p {
		if !p.byz {
			needed[pid] = struct{}{}
		}
	}
	c.pLock.Unlock()

	i := 0
	for {
		c.dLock.Lock()
		for pid := range c.deliverMap[uid] {
			delete(needed, pid)
		}
		c.dLock.Unlock()

		if len(needed) == 0 {
			return c.aggregateStats(uid)
		}

		if i == 0 {
			fmt.Printf("waiting for %v more (%v) delivers: %v\n", len(needed), uid, needed)
		}
		i = (i + 1) % 5

		time.Sleep(c.cfg.PollDelay * 2)
	}
}

func (c *Controller) aggregateStats(uid uint32) Stats {
	c.pLock.Lock()
	defer c.pLock.Unlock()

	latency := time.Duration(0)
	cnt := 0
	recv := 0
	bdMerged := 0
	minRecv, maxRecv := -1, -1
	transmitted := 0
	dMerged := 0
	pMerged := 0

	for _, p := range c.p {
		s := p.p.Stats()
		del := s.Deliveries[uid]
		c.dLock.Lock()
		lat := del.Sub(c.sendMap[uid])
		c.dLock.Unlock()

		if lat > latency {
			latency = lat
		}

		rec := s.Relayed[uid]
		bdMerged += s.BDMerged[uid]
		recv += rec
		cnt += s.MsgSent[uid]
		dMerged += s.DMerged[uid]
		transmitted += int(s.BytesTransmitted[uid])
		pMerged += s.PayloadsMerged[uid]

		if rec > maxRecv {
			maxRecv = rec
		} else if rec < minRecv || minRecv == -1 {
			minRecv = rec
		}
	}

	return Stats{
		Latency:          latency,
		MsgCount:         cnt,
		RelayCnt:         recv,
		MeanRelayCount:   float64(recv) / float64(len(c.p)-1),
		MinRelayCnt:      minRecv,
		MaxRelayCnt:      maxRecv,
		BDMessagedMerged: bdMerged,
		BytesTransmitted: transmitted,
		DMessagesMerged:  dMerged,
		PayloadsMerged:   pMerged,
	}
}

func (c *Controller) Close() {
	close(c.stopCh)

	for _, ch := range c.channels {
		close(ch)
	}
}

func (c *Controller) send(id uint64, t uint8, b interface{}) {
	c.channels[id] <- process.Message{
		Ctl:  true,
		Type: t,
		Data: b,
	}
}

func (c *Controller) run() {
	for {
		select {
		case <-c.stopCh:
			close(c.ctl)
			return
		default:
		}

		m := <-c.ctl
		c.handleMsg(m.Src, m.Type, m.Data)
	}
}

func (c *Controller) handleMsg(src uint64, t uint8, b interface{}) {
	//fmt.Printf("server got data from %v (type=%v): %v\n", src, t, b)

	switch t {
	case msg.RunnerAliveType:
		r := b.(msg.RunnerStatus)
		//if err := r.Decode(b); err != nil {
		//	fmt.Printf("failed to decode msg: %v\n", err)
		//	return
		//}

		c.pLock.Lock()
		p := c.p[r.ID]
		p.alive = true
		c.p[r.ID] = p
		c.al += 1
		c.pLock.Unlock()

		//fmt.Printf("runner %v is alive (%v)\n", src, c.al)
	case msg.RunnerReadyType:
		r := b.(msg.RunnerStatus)
		//if err := r.Decode(b); err != nil {
		//	fmt.Printf("failed to decode msg: %v\n", err)
		//	return
		//}

		c.pLock.Lock()
		p := c.p[r.ID]
		p.ready = true
		c.p[r.ID] = p
		c.rdy += 1
		c.pLock.Unlock()

		//fmt.Printf("runner %v is ready (%v/%v)\n", src, c.rdy, len(c.p))
	case msg.RunnerFailedType:
		r := b.(msg.RunnerFailure)
		//if err := r.Decode(b); err != nil {
		//	fmt.Printf("failed to decode msg: %v\n", err)
		//	return
		//}

		c.pLock.Lock()
		p := c.p[r.ID]
		p.err = r.Err
		c.p[r.ID] = p
		c.pLock.Unlock()

		//fmt.Printf("runner %v failed: %v\n", r.ID, r.Err)
	case msg.MessageDeliveredType:
		r := b.(msg.MessageDelivered)
		//if err := r.Decode(b); err != nil {
		//	fmt.Printf("failed to decode msg: %v\n", err)
		//	return
		//}

		c.dLock.Lock()
		if !reflect.DeepEqual(r.Payload, c.payloadMap[r.Id]) {
			fmt.Printf("process %v delived invalid payload, BRB guarantees violated: got %v, wanted %v\n",
				src, r.Payload, c.payloadMap[r.Id])
			os.Exit(1)
		}

		c.deliverMap[r.Id][src] = struct{}{}
		//del := len(c.deliverMap[r.Id])
		c.dLock.Unlock()

		c.pLock.Lock()
		//fmt.Printf("runner %v has delivered %v (%v/%v-F)\n", src, r.Id, del, len(c.p))
		c.pLock.Unlock()
	}
}
