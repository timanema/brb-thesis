package ctrl

import (
	"encoding/binary"
	"fmt"
	"github.com/pebbe/zmq4"
	"github.com/pkg/errors"
	"rp-runner/msg"
	"rp-runner/process"
	"sync"
	"time"
)

const pollInterval = time.Millisecond * 300

type proc struct {
	p            *process.Process
	alive, ready bool

	err error
}

type ControllerInfo struct {
	ID, Sock string
}

type Controller struct {
	s      *zmq4.Socket
	stopCh chan struct{}

	p     map[uint16]proc
	pLock sync.Mutex
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

	c := &Controller{s: s, stopCh: make(chan struct{}), p: make(map[uint16]proc)}
	go c.run()

	return c, nil
}

func (c *Controller) StartProcess(id uint16, cfg process.Config, neighbours []uint16) error {
	c.pLock.Lock()
	defer c.pLock.Unlock()

	p, err := process.StartProcess(id, cfg, c.stopCh, neighbours)
	if err != nil {
		return errors.Wrap(err, "unable to start process")
	}

	c.p[id] = proc{p: p}
	return nil
}

// TODO: tack on statistics etc later
func (c *Controller) TriggerMessageSend(id uint16, payload []byte) error {
	m := &msg.TriggerMessage{Payload: payload}
	b, err := m.Encode()
	if err != nil {
		return errors.Wrap(err, "failed to encode payload message")
	}

	return errors.Wrapf(c.send(id, msg.TriggerMessageType, b), "failed to send message to %v", id)
}

func (c *Controller) WaitForAlive() error {
	waiting := true
	for waiting {
		waiting = false
		for pic, p := range c.p {
			if p.err != nil {
				return errors.Wrapf(p.err, "process %v failed", pic)
			} else if !p.alive {
				fmt.Printf("waiting for %v alive\n", pic)
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
				fmt.Printf("waiting for %v ready\n", pic)
				waiting = true
				time.Sleep(pollInterval)
				break
			}
		}
	}

	return nil
}

func (c *Controller) Close() {
	close(c.stopCh)
}

func (c *Controller) send(id uint16, t uint8, b []byte) error {
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
			c.handleMsg(binary.BigEndian.Uint16(m[0]), m[1][0], m[2])
		} else {
			fmt.Printf("discarding bogus message: %v\n", m)
		}
	}
}

func (c *Controller) handleMsg(src uint16, t uint8, b []byte) {
	fmt.Printf("server got data from %v (type=%v): %v\n", src, t, b)

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

		fmt.Printf("runner %v is alive\n", r.ID)
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

		fmt.Printf("runner %v is ready\n", r.ID)
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
	}
}
