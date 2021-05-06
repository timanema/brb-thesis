package ctrl

import (
	"encoding/binary"
	"fmt"
	"github.com/pebbe/zmq4"
	"github.com/pkg/errors"
	"rp-runner/msg"
)

type ControllerInfo struct {
	ID, Sock string
}

type Controller struct {
	s      *zmq4.Socket
	stopCh chan struct{}
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

	c := &Controller{s: s, stopCh: make(chan struct{})}
	go c.run()

	return c, nil
}

func (c *Controller) Close() {
	close(c.stopCh)
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
		var r msg.RunnerAlive
		if err := r.Decode(b); err != nil {
			fmt.Printf("failed to decode msg: %v\n", err)
			return
		}

		fmt.Printf("runner %v is alive\n", r.ID)
	case msg.RunnerReadyType:
		var r msg.RunnerReady
		if err := r.Decode(b); err != nil {
			fmt.Printf("failed to decode msg: %v\n", err)
			return
		}

		fmt.Printf("runner %v is ready\n", r.ID)
	}
}
