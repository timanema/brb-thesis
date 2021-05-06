package process

import (
	"fmt"
	"github.com/pebbe/zmq4"
	"github.com/pkg/errors"
	"rp-runner/msg"
	"time"
)

type Config struct {
	CtrlID, CtrlSock string
	Sock             string
	MaxRetries       int
	RetryDelay       time.Duration
}

type Process struct {
	id     uint16
	ctrl   *zmq4.Socket
	cfg    Config
	stopCh <-chan struct{}
}

func StartProcess(id uint16, cfg Config, stopCh <-chan struct{}) (*Process, error) {
	s, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create ZeroMQ context")
	}

	if err := s.SetRouterMandatory(1); err != nil {
		return nil, errors.Wrap(err, "unable to set mandatory routing flag")
	}

	if err := s.SetIdentity(idToString(id)); err != nil {
		return nil, errors.Wrap(err, "unable to set ZeroMQ identity")
	}

	if err := s.Connect(cfg.CtrlSock); err != nil {
		return nil, errors.Wrapf(err, "unable to connect to socket %v", cfg.CtrlID)
	}

	if err := s.Bind(fmt.Sprintf(cfg.Sock, id)); err != nil {
		return nil, errors.Wrapf(err, "unable to bind to socket %v", fmt.Sprintf(cfg.Sock, id))
	}

	p := &Process{id: id, ctrl: s, cfg: cfg, stopCh: stopCh}

	go p.run()

	if err := p.signalAlive(); err != nil {
		return nil, errors.Wrap(err, "unable to communicate with controller")
	}

	return p, nil
}

func (p *Process) signalAlive() error {
	m := &msg.RunnerAlive{ID: p.id}
	b, err := m.Encode()
	if err != nil {
		return errors.Wrap(err, "unable to encode alive message")
	}

	retries := 0

	for {
		_, err = p.ctrl.SendMessage(p.cfg.CtrlID, []byte{msg.RunnerAliveType}, b)
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

		m, err := p.ctrl.RecvBytes(0)

		if err != nil {
			fmt.Printf("err while reading: %v\n", err)
		} else {
			fmt.Printf("process %v got data: %v\n", p.id, m)
		}
	}
}
