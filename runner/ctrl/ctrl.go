package ctrl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/pebbe/zmq4"
	"github.com/pkg/errors"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"math/rand"
	"os"
	"reflect"
	"rp-runner/brb"
	"rp-runner/msg"
	"rp-runner/process"
	"sync"
	"time"
)

const pollInterval = time.Millisecond * 300

type proc struct {
	p                 *process.Process
	alive, ready, byz bool

	err error
}

type ControllerInfo struct {
	ID, Sock string
}

type Controller struct {
	s      *zmq4.Socket
	stopCh chan struct{}

	p     map[uint64]proc
	pLock sync.Mutex

	payloadMap map[uint32][]byte
	deliverMap map[uint32]map[uint64]struct{}
	sendMap    map[uint32]time.Time
	dLock      sync.Mutex
}

func StartController(info ControllerInfo) (*Controller, error) {
	s, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create ZeroMQ context")
	}

	if err := s.SetRouterMandatory(1); err != nil {
		return nil, errors.Wrap(err, "unable to set mandatory routing flag")
	}

	if err := s.SetIdentity(info.ID); err != nil {
		return nil, errors.Wrap(err, "unable to set ZeroMQ identity")
	}

	if err := s.Bind(info.Sock); err != nil {
		return nil, errors.Wrapf(err, "unable to bind to socket %v", info.Sock)
	}

	c := &Controller{s: s,
		stopCh:     make(chan struct{}),
		p:          make(map[uint64]proc),
		payloadMap: make(map[uint32][]byte),
		deliverMap: make(map[uint32]map[uint64]struct{}),
		sendMap:    make(map[uint32]time.Time),
	}
	go c.run()

	return c, nil
}

func (c *Controller) StartProcess(cfg process.Config, bp brb.Protocol) error {
	p, err := process.StartProcess(cfg.ByzConfig.Id, cfg, c.stopCh, cfg.ByzConfig.Neighbours, bp)
	if err != nil {
		return errors.Wrap(err, "unable to start process")
	}

	c.pLock.Lock()
	c.p[cfg.ByzConfig.Id] = proc{p: p, byz: cfg.ByzConfig.Byz}
	c.pLock.Unlock()
	return nil
}

func (c *Controller) contains(n uint64, xs []uint64) bool {
	for _, v := range xs {
		if n == v {
			return true
		}
	}

	return false
}

// TODO: random byzantine nodes?
func (c *Controller) StartProcesses(cfg process.Config, g graph.WeightedUndirected, bp brb.Protocol, F int, byzEx []uint64) error {
	byzLeft := F
	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()
		to := g.From(n.ID())
		neighbours := make([]uint64, 0, to.Len())

		for to.Next() {
			neighbours = append(neighbours, uint64(to.Node().ID()))
		}

		pg := simple.NewWeightedUndirectedGraph(0, 0)
		graph.CopyWeighted(pg, g)

		byz := byzLeft > 0 && !c.contains(uint64(n.ID()), byzEx)
		if byz {
			byzLeft -= 1
		}

		pcfg := cfg
		pcfg.ByzConfig = brb.Config{
			Byz:           byz,
			F:             F,
			Id:            uint64(n.ID()),
			Neighbours:    neighbours,
			Graph:         pg,
			KnownTopology: true,
		}

		if err := c.StartProcess(pcfg, reflect.New(reflect.ValueOf(bp).Elem().Type()).Interface().(brb.Protocol)); err != nil {
			return errors.Wrap(err, "failed to start process")
		}
	}

	return nil
}

// TODO: tack on statistics etc later
func (c *Controller) TriggerMessageSend(id uint64, payload []byte) (uint32, error) {
	uid := rand.Uint32()

	m := &msg.TriggerMessage{Id: uid, Payload: payload}
	b, err := m.Encode()
	if err != nil {
		return 0, errors.Wrap(err, "failed to encode payload message")
	}

	c.dLock.Lock()
	c.payloadMap[uid] = payload
	c.deliverMap[uid] = make(map[uint64]struct{})
	c.sendMap[uid] = time.Now()
	c.dLock.Unlock()

	return uid, errors.Wrapf(c.send(id, msg.TriggerMessageType, b), "failed to send message to %v", id)
}

func (c *Controller) WaitForAlive() error {
	waiting := true
	for waiting {
		waiting = false
		for pic, p := range c.p {
			if p.err != nil {
				return errors.Wrapf(p.err, "process %v failed", pic)
			} else if !p.alive {
				//fmt.Printf("waiting for %v alive\n", pic)
				waiting = true
				time.Sleep(pollInterval)
				break
			}
		}
	}

	return nil
}

func (c *Controller) WaitForReady() error {
	waiting := true
	for waiting {
		waiting = false
		for pic, p := range c.p {
			if p.err != nil {
				return errors.Wrapf(p.err, "process %v failed", pic)
			} else if !p.ready {
				//fmt.Printf("waiting for %v ready\n", pic)
				waiting = true
				time.Sleep(pollInterval)
				break
			}
		}
	}

	return nil
}

func (c *Controller) WaitForDeliver(uid uint32) Stats {
	// TODO: use ids of normal nodes instead of len check?
	c.pLock.Lock()
	needed := 0
	for _, p := range c.p {
		if !p.byz {
			needed += 1
		}
	}
	c.pLock.Unlock()

	for {
		c.dLock.Lock()
		del := len(c.deliverMap[uid])
		c.dLock.Unlock()

		if del >= needed {
			return c.aggregateStats(uid)
		}
		time.Sleep(pollInterval * 2)
	}
}

func (c *Controller) aggregateStats(uid uint32) Stats {
	c.pLock.Lock()
	c.dLock.Lock()
	defer c.dLock.Unlock()
	defer c.pLock.Unlock()

	latency := time.Duration(0)
	cnt := 0

	for _, p := range c.p {
		s := p.p.Stats()
		del := s.Deliveries[uid]
		lat := del.Sub(c.sendMap[uid])

		if lat > latency {
			latency = lat
		}

		cnt += s.MsgSent[uid]
	}

	return Stats{
		Latency:  latency,
		MsgCount: cnt,
	}
}

func (c *Controller) Close() {
	close(c.stopCh)
}

func (c *Controller) send(id uint64, t uint8, b []byte) error {
	_, err := c.s.SendMessage(process.IdToString(id), []byte{t}, b, []byte{process.ControlIdMagic})

	return err
}

func (c *Controller) run() {
	for {
		select {
		case <-c.stopCh:
			_ = c.s.Close()
			return
		default:
		}

		m, err := c.s.RecvMessageBytes(0)

		if err != nil {
			fmt.Printf("err while reading: %v\n", err)
		} else if len(m) >= 3 {
			c.handleMsg(binary.BigEndian.Uint64(m[0]), m[1][0], m[2])
		} else {
			fmt.Printf("discarding bogus message: %v\n", m)
		}
	}
}

func (c *Controller) handleMsg(src uint64, t uint8, b []byte) {
	//fmt.Printf("server got data from %v (type=%v): %v\n", src, t, b)

	switch t {
	case msg.RunnerAliveType:
		var r msg.RunnerStatus
		if err := r.Decode(b); err != nil {
			fmt.Printf("failed to decode msg: %v\n", err)
			return
		}

		c.pLock.Lock()
		p := c.p[r.ID]
		p.alive = true
		c.p[r.ID] = p
		c.pLock.Unlock()

		//fmt.Printf("runner %v is alive\n", r.ID)
	case msg.RunnerReadyType:
		var r msg.RunnerStatus
		if err := r.Decode(b); err != nil {
			fmt.Printf("failed to decode msg: %v\n", err)
			return
		}

		c.pLock.Lock()
		p := c.p[r.ID]
		p.ready = true
		c.p[r.ID] = p
		c.pLock.Unlock()

		//fmt.Printf("runner %v is ready\n", r.ID)
	case msg.RunnerFailedType:
		var r msg.RunnerFailure
		if err := r.Decode(b); err != nil {
			fmt.Printf("failed to decode msg: %v\n", err)
			return
		}

		c.pLock.Lock()
		p := c.p[r.ID]
		p.err = r.Err
		c.p[r.ID] = p
		c.pLock.Unlock()

		fmt.Printf("runner %v failed: %v\n", r.ID, r.Err)
	case msg.MessageDeliveredType:
		var r msg.MessageDelivered
		if err := r.Decode(b); err != nil {
			fmt.Printf("failed to decode msg: %v\n", err)
			return
		}

		c.dLock.Lock()
		if bytes.Compare(r.Payload, c.payloadMap[r.Id]) != 0 {
			fmt.Printf("process %v delived invalid payload, BRB guarantees violated: got %v, wanted %v\n",
				src, r.Payload, c.payloadMap[r.Id])
			os.Exit(1)
		}

		c.deliverMap[r.Id][src] = struct{}{}
		c.dLock.Unlock()
	}
}
