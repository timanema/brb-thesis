package msg

import (
	"bytes"
	"encoding/gob"
	"github.com/pkg/errors"
)

const RunnerAliveType uint8 = 1
const RunnerReadyType uint8 = 2

type RunnerAlive struct {
	ID uint16
}

func (r *RunnerAlive) Encode() ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0, 4))
	enc := gob.NewEncoder(b)
	err := enc.Encode(r)

	return b.Bytes(), errors.Wrap(err, "unable to encode message")
}

func (r *RunnerAlive) Decode(b []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	return errors.Wrap(dec.Decode(r), "unable to decode message")
}

type RunnerReady struct {
	ID uint16
}

func (r *RunnerReady) Encode() ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0, 4))
	enc := gob.NewEncoder(b)
	err := enc.Encode(r)

	return b.Bytes(), errors.Wrap(err, "unable to encode message")
}

func (r *RunnerReady) Decode(b []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	return errors.Wrap(dec.Decode(r), "unable to decode message")
}
