package msg

import (
	"bytes"
	"encoding/gob"
	"github.com/pkg/errors"
)

const RunnerAliveType uint8 = 1
const RunnerReadyType uint8 = 2
const RunnerFailedType uint8 = 3

type RunnerStatus struct {
	ID uint16
}

func (r *RunnerStatus) Encode() ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0, 4))
	enc := gob.NewEncoder(b)
	err := enc.Encode(r)

	return b.Bytes(), errors.Wrap(err, "unable to encode message")
}

func (r *RunnerStatus) Decode(b []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	return errors.Wrap(dec.Decode(r), "unable to decode message")
}

type RunnerFailure struct {
	ID  uint16
	Err error
}

func (r *RunnerFailure) Encode() ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0, 4))
	enc := gob.NewEncoder(b)
	err := enc.Encode(r)

	return b.Bytes(), errors.Wrap(err, "unable to encode message")
}

func (r *RunnerFailure) Decode(b []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	return errors.Wrap(dec.Decode(r), "unable to decode message")
}