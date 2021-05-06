package msg

import (
	"bytes"
	"encoding/gob"
	"github.com/pkg/errors"
)

const TriggerMessageType uint8 = 5
const WrapperDataType uint8 = 6
const MessageDeliveredType uint8 = 7

type TriggerMessage struct {
	Id      uint32
	Payload []byte
}

func (r *TriggerMessage) Encode() ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0, 20))
	enc := gob.NewEncoder(b)
	err := enc.Encode(r)

	return b.Bytes(), errors.Wrap(err, "unable to encode message")
}

func (r *TriggerMessage) Decode(b []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	return errors.Wrap(dec.Decode(r), "unable to decode message")
}

type WrapperDataMessage struct {
	T    uint8
	Id   uint32
	Data []byte
}

func (r *WrapperDataMessage) Encode() ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0, 20))
	enc := gob.NewEncoder(b)
	err := enc.Encode(r)

	return b.Bytes(), errors.Wrap(err, "unable to encode message")
}

func (r *WrapperDataMessage) Decode(b []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	return errors.Wrap(dec.Decode(r), "unable to decode message")
}

type MessageDelivered struct {
	Id      uint32
	Payload []byte
}

func (r *MessageDelivered) Encode() ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0, 20))
	enc := gob.NewEncoder(b)
	err := enc.Encode(r)

	return b.Bytes(), errors.Wrap(err, "unable to encode message")
}

func (r *MessageDelivered) Decode(b []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	return errors.Wrap(dec.Decode(r), "unable to decode message")
}