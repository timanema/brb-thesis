package process

import (
	"encoding/binary"
	"fmt"
	"github.com/pebbe/zmq4"
	"github.com/pkg/errors"
	"os"
	"rp-runner/msg"
	"time"
)

type Config struct {
	CtrlID, CtrlSock           string
	Sock                       string
	MaxRetries                 int
	RetryDelay, NeighbourDelay time.Duration
}

type neighbour struct {
	s     *zmq4.Socket
	ready bool
}

type Process struct {
	id     uint16
	s      *zmq4.Socket
	cfg    Config
	stopCh <-chan struct{}

	neighbours map[uint16]neighbour
}

func StartProcess(id uint16, cfg Config, stopCh <-chan struct{}, neighbours []uint16) (*Process, error) {
	s, err := createSocket(IdToString(id), cfg.CtrlSock)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create socket")
	}

	if err := s.Bind(fmt.Sprintf(cfg.Sock, id)); err != nil {
		return nil, errors.Wrapf(err, "unable to bind to socket %v", fmt.Sprintf(cfg.Sock, id))
	}

	// Make a map indicating readiness
	nmap := make(map[uint16]neighbour, len(neighbours))
	for _, n := range neighbours {
		nsoc, err := createSocket(IdToString(id), fmt.Sprintf(cfg.Sock, n))
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create socket to neighbour %v", n)
		}

		nmap[n] = neighbour{s: nsoc}
	}

	p := &Process{id: id, s: s, cfg: cfg, stopCh: stopCh, neighbours: nmap}

	go p.run()

	if err := p.waitForConnection(); err != nil {
		return nil, errors.Wrap(err, "unable to communicate with controller")
	}

	go p.checkNeighbours()

	return p, nil
}

func createSocket(id, endpoint string) (*zmq4.Socket, error) {
	s, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create ZeroMQ socket")
	}

	if err := s.SetRouterMandatory(1); err != nil {
		return nil, errors.Wrap(err, "unable to set mandatory routing flag")
	}

	if err := s.SetIdentity(id); err != nil {
		return nil, errors.Wrap(err, "unable to set ZeroMQ identity")
	}

	if err := s.Connect(endpoint); err != nil {
		return nil, errors.Wrapf(err, "unable to connect to socket %v", endpoint)
	}

	return s, nil
}

func (p *Process) checkNeighbours() {
	m := &msg.RunnerStatus{ID: p.id}
	b, err := m.Encode()
	if err != nil {
		errMsg := &msg.RunnerFailure{ID: p.id, Err: errors.Wrap(err, "unable to encode ready message")}
		if b, err = errMsg.Encode(); err != nil {
			fmt.Printf("process %v is unable to encode failure message: %v\n", p.id, err)
			os.Exit(1)
		}

		if _, err = p.s.SendMessage(p.cfg.CtrlID, []byte{msg.RunnerFailedType}, b); err != nil {
			fmt.Printf("process %v is unable to send failure message: %v\n", p.id, err)
			os.Exit(1)
		}
	}

	waiting := true
	for waiting {
		select {
		case <-p.stopCh:
			return
		default:
		}

		waiting = false
		for nid, n := range p.neighbours {
			if !n.ready {
				_, err = n.s.SendMessage(IdToString(nid), []byte{msg.RunnerReadyType}, b)
				if err != nil {
					waiting = true
					time.Sleep(p.cfg.NeighbourDelay)
				} else {
					n.ready = true
					p.neighbours[nid] = n
				}
			}
		}
	}

	_, err = p.s.SendMessage(p.cfg.CtrlID, []byte{msg.RunnerReadyType}, b)
}

func (p *Process) waitForConnection() error {
	m := &msg.RunnerStatus{ID: p.id}
	b, err := m.Encode()
	if err != nil {
		return errors.Wrap(err, "unable to encode alive message")
	}

	retries := 0

	for {
		_, err = p.s.SendMessage(p.cfg.CtrlID, []byte{msg.RunnerAliveType}, b)
		if err != nil {
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

func (p *Process) run() {
	for {
		select {
		case <-p.stopCh:
			return
		default:
		}

		m, err := p.s.RecvMessageBytes(0)

		if err != nil {
			fmt.Printf("err while reading: %v\n", err)
		} else if len(m) >= 3 {
			p.handleMsg(binary.BigEndian.Uint16(m[0]), m[1][0], m[2], len(m) >= 4 && m[3][0] == ControlIdMagic)
		} else {
			fmt.Printf("discarding bogus message: %v\n", m)
		}
	}
}

func (p *Process) handleMsg(src uint16, t uint8, b []byte, ctrl bool) {
	if ctrl {
		fmt.Printf("process %v got data from %v (type=%v): %v\n", p.id, p.cfg.CtrlID, t, b)
	} else {
		fmt.Printf("process %v got data from %v (type=%v): %v\n", p.id, src, t, b)
	}
}

func (p *Process) send(id uint16, t uint8, b []byte, ctrl bool) error {
	if ctrl {
		_, err := p.s.SendMessage(p.cfg.CtrlID, []byte{t}, b)
		return err
	}

	n, ok := p.neighbours[id]
	if !ok {
		return errors.Errorf("unknown neighbour id %v", id)
	}

	_, err := n.s.SendMessage(IdToString(id), []byte{t}, b)

	return err
}

// Adding abstraction for BRB protocols
func (p *Process) Deliver(payload []byte) {
	fmt.Printf("process %v got delivered: %v", p.id, string(payload))

	// TODO: notify controller
}

func (p *Process) Send(t uint8, dest uint16, data []byte) {
	fmt.Printf("process %v is sending %v bytes (type=%v) to %v\n", p.id, len(data), t, dest)

	m := &msg.WrapperDataMessage{
		T:    t,
		Src:  p.id,
		Data: data,
	}
	b, err := m.Encode()
	if err != nil {
		// TODO: send to controller
		fmt.Printf("process %v failed to encode wrapper data message: %v\n", p.id, err)
		os.Exit(1)
	}

	err = p.send(dest, msg.WrapperDataType, b, false)
	if err != nil {
		// TODO: send to controller
		fmt.Printf("process %v failed to send wrapper data message: %v\n", p.id, err)
		os.Exit(1)
	}
}
