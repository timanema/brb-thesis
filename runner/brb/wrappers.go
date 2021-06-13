package brb

import (
	"reflect"
	"rp-runner/brb/algo"
)

type BrachaDolevWrapperMsg struct {
	Msgs            []BrachaDolevMessage
	OriginalSrc     uint64
	OriginalId      uint32
	OriginalPayload Size
}

func (b BrachaDolevWrapperMsg) SizeOf() uintptr {
	r := uintptr(0)
	for _, msg := range b.Msgs {
		r += msg.SizeOf()
	}

	return r + reflect.TypeOf(b.OriginalSrc).Size() + reflect.TypeOf(b.OriginalId).Size() + b.OriginalPayload.SizeOf()
}

// Unpack creates all DolevKnownImprovedMessage instances from the current BrachaDolevWrapper
func (b BrachaDolevWrapperMsg) Unpack() []DolevKnownImprovedMessage {
	res := make([]DolevKnownImprovedMessage, 0, len(b.Msgs))

	for _, msg := range b.Msgs {
		bm := BrachaMessage{
			Src:     b.OriginalSrc,
			Id:      b.OriginalId,
			Payload: b.OriginalPayload,
		}
		dm := DolevKnownImprovedMessage{
			Src: msg.Src,
			Id:  msg.Id,
			Payload: brachaWrapper{
				messageType: msg.Type,
				msg:         bm,
			},
			Paths:   msg.Paths,
			Partial: msg.Partial,
		}

		res = append(res, dm)
	}

	return res
}

// Pack fills the current BrachaDolevWrapper with information from a slice of DolevKnownImprovedMessage instances
func Pack(original []DolevKnownImprovedMessage) BrachaDolevWrapperMsg {
	msgs := make([]BrachaDolevMessage, 0, len(original))
	bdw := BrachaDolevWrapperMsg{}

	for _, msg := range original {
		bw := msg.Payload.(brachaWrapper)

		msgs = append(msgs, BrachaDolevMessage{
			Src:     msg.Src,
			Id:      msg.Id,
			Type:    bw.messageType,
			Paths:   msg.Paths,
			Partial: msg.Partial,
		})

		bm := bw.msg.(BrachaMessage)
		bdw.OriginalSrc = bm.Src
		bdw.OriginalId = bm.Id
		bdw.OriginalPayload = bm.Payload
	}

	bdw.Msgs = msgs
	return bdw
}

// TODO: better name
type dolevWrapperWrapper struct {
	Src        uint64
	Id         uint32
	TrackingId uint32
	Paths      []algo.DolevPath
}

func (d dolevWrapperWrapper) SizeOf() uintptr {
	return reflect.TypeOf(d.Src).Size() + reflect.TypeOf(d.Id).Size() +
		reflect.TypeOf(d.TrackingId).Size() + algo.SizeOfMultiplePaths(d.Paths)
}

type DolevWrapperMessage struct {
	Msgs    []dolevWrapperWrapper
	Payload Size
}

func (d DolevWrapperMessage) SizeOf() uintptr {
	r := uintptr(0)
	for _, msg := range d.Msgs {
		r += msg.SizeOf()
	}

	return r + d.Payload.SizeOf()
}

func (d DolevWrapperMessage) Unpack(m DolevKnownImprovedMessage, uid uint32) ([]DolevKnownImprovedMessage, []uint32) {
	res := make([]DolevKnownImprovedMessage, 0, len(d.Msgs)+1)
	tracking := make([]uint32, 0, len(d.Msgs)+1)

	for _, m := range d.Msgs {
		res = append(res, DolevKnownImprovedMessage{
			Src:     m.Src,
			Id:      m.Id,
			Payload: d.Payload,
			Paths:   m.Paths,
		})
		tracking = append(tracking, m.TrackingId)
	}

	res = append(res, DolevKnownImprovedMessage{
		Src:     m.Src,
		Id:      m.Id,
		Payload: d.Payload,
		Paths:   m.Paths,
	})
	tracking = append(tracking, uid)

	return res, tracking
}
