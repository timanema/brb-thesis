package process

import (
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"os"
	"rp-runner/brb"
	"rp-runner/msg"
	"sync"
	"time"
)

type Config struct {
	MaxRetries                 int
	RetryDelay, NeighbourDelay time.Duration
	ByzConfig                  brb.Config
}

type Stats struct {
	Deliveries map[uint32]time.Time
	MsgSent    map[uint32]int
}

type Process struct {
	channels map[uint64]chan Message
	ctl      chan Message
	flushing *atomic.Bool

	id     uint64
	cfg    Config
	stopCh <-chan struct{}

	stats Stats
	sLock sync.Mutex

	brb brb.Protocol

	neighbours map[uint64]bool
}

func StartProcess(id uint64, cfg Config, stopCh <-chan struct{}, neighbours []uint64, brb brb.Protocol, ctl chan Message) (*Process, error) {
	nmap := make(map[uint64]bool, len(neighbours))
	for _, n := range neighbours {
		nmap[n] = false
	}

	stats := Stats{Deliveries: make(map[uint32]time.Time), MsgSent: make(map[uint32]int)}
	p := &Process{ctl: ctl, flushing: atomic.NewBool(false), id: id, cfg: cfg, stopCh: stopCh, stats: stats, brb: brb, neighbours: nmap}

	return p, nil
}

func (p *Process) Start(channels map[uint64]chan Message) error {
	p.channels = channels

	if err := p.waitForConnection(); err != nil {
		return errors.Wrap(err, "unable to communicate with controller")
	}

	go p.run()

	go func() {
		//fmt.Printf("init %v\n", p.id)
		p.brb.Init(p, p, p.cfg.ByzConfig)
		//fmt.Printf("done init %v\n", p.id)

		p.checkNeighbours()
	}()

	return nil
}

func (p *Process) send(id uint64, t uint8, b interface{}, ctrl bool) error {
	m := Message{
		Src:  p.id,
		Type: t,
		Data: b,
	}

	if ctrl {
		p.ctl <- m
	} else if !p.flushing.Load() {
		c, ok := p.channels[id]
		if !ok {
			return errors.Errorf("proc %v is not connected to %v", p.id, id)
		}

		select {
		case c <- m:
			break
		default:
			// Allows runs to go over proc limit, at the expensive of exploding memory usage
			go func() {
				c <- m
			}()
		}
	}

	return nil
}

func (p *Process) checkNeighbours() {
	m := msg.RunnerStatus{ID: p.id}

	waiting := true
	for waiting {
		select {
		case <-p.stopCh:
			return
		default:
		}

		waiting = false
		for nid, n := range p.neighbours {
			if !n {
				if err := p.send(nid, msg.RunnerPingType, []byte{0x00}, false); err != nil {
					//fmt.Printf("proc %v got err to %v: %v\n", p.id, nid, err)
					waiting = true
					time.Sleep(p.cfg.NeighbourDelay)
				} else {
					p.neighbours[nid] = true
				}
			}
		}
	}

	if err := p.send(0, msg.RunnerReadyType, m, true); err != nil {
		fmt.Printf("process %v is unable to send ready message: %v\n", p.id, err)
		os.Exit(1)
	}
}

func (p *Process) waitForConnection() error {
	m := msg.RunnerStatus{ID: p.id}

	retries := 0

	for {
		if err := p.send(0, msg.RunnerAliveType, m, true); err != nil {
			// TODO: better error matching
			if err.Error() == "no route to host" {
				//fmt.Printf("socket not yet ready: %v\n", err)
				time.Sleep(p.cfg.RetryDelay)
				retries += 1

				if retries >= p.cfg.MaxRetries {
					return errors.Errorf("failed to send alive message to controller after %v tries", p.cfg.MaxRetries)
				}

				continue
			} else {
				return errors.Wrap(err, "unable to send ready message to controller")
			}
		}

		return nil
	}
}

func (p *Process) Flush() {
	p.flushing.Store(true)

	go func() {
		for {
			select {
			case <-p.channels[p.id]:
				continue
			default:
				if !p.flushing.Load() {
					return
				}
				time.Sleep(time.Millisecond * 200)
			}
		}
	}()
}

func (p *Process) StopFlush() {
	p.flushing.Store(false)
}

func (p *Process) run() {
	for {
		select {
		case <-p.stopCh:
			return
		default:
		}

		m := <-p.channels[p.id]
		if p.flushing.Load() {
			continue
		}

		p.handleMsg(m.Src, m.Type, m.Data, m.Ctl)
	}
}

func (p *Process) handleMsg(src uint64, t uint8, b interface{}, ctrl bool) {
	//if ctrl {
	//	fmt.Printf("process %v got data from %v (type=%v): %v\n", p.id, p.cfg.CtrlID, t, b)
	//} else {
	//	fmt.Printf("process %v got data from %v (type=%v): %v\n", p.id, src, t, b)
	//}

	switch t {
	case msg.WrapperDataType:
		r := b.(msg.WrapperDataMessage)

		p.brb.Receive(r.T, src, r.Id, r.Data)
	case msg.TriggerMessageType:
		r := b.(msg.TriggerMessage)

		p.stats.MsgSent[r.Id] = 0
		p.brb.Send(r.Id, r.Payload)
	}
}

// Adding abstraction for BRB protocols
func (p *Process) Deliver(uid uint32, payload []byte) {
	//fmt.Printf("process %v got delivered (%v): %v\n", p.id, uid, string(payload))

	m := msg.MessageDelivered{
		Id:      uid,
		Payload: payload,
	}

	err := p.send(0, msg.MessageDeliveredType, m, true)
	if err != nil {
		fmt.Printf("process %v failed to send deliver message: %v\n", p.id, err)
		os.Exit(1)
	}

	p.sLock.Lock()
	p.stats.Deliveries[uid] = time.Now()
	p.sLock.Unlock()
}

func (p *Process) Send(t uint8, dest uint64, uid uint32, data interface{}) {
	//fmt.Printf("process %v is sending %v bytes (type=%v, id=%v) to %v\n", p.id, len(data), t, uid, dest)

	m := msg.WrapperDataMessage{
		T:    t,
		Id:   uid,
		Data: data,
	}

	err := p.send(dest, msg.WrapperDataType, m, false)
	if err != nil {
		fmt.Printf("process %v failed to send wrapper data message to %v: %v\n", p.id, dest, err)
		os.Exit(1)
	}

	p.sLock.Lock()
	p.stats.MsgSent[uid] += 1
	p.sLock.Unlock()
}

func (p *Process) Stats() Stats {
	p.sLock.Lock()
	defer p.sLock.Unlock()

	s := p.stats
	return s
}
