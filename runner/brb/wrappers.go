package brb

type BrachaDolevWrapperMsg struct {
	Msgs            []BrachaDolevMessage
	OriginalSrc     uint64
	OriginalId      uint32
	OriginalPayload interface{}
	Included        []uint64
}

// Unpack creates all DolevKnownImprovedMessage instances from the current BrachaDolevWrapper
func (b BrachaDolevWrapperMsg) Unpack() []DolevKnownImprovedMessage {
	res := make([]DolevKnownImprovedMessage, 0, len(b.Msgs))

	for _, msg := range b.Msgs {
		bm := BrachaImprovedMessage{
			BrachaMessage: BrachaMessage{
				Src:     b.OriginalSrc,
				Id:      b.OriginalId,
				Payload: b.OriginalPayload,
			},
			Included: b.Included,
		}
		dm := DolevKnownImprovedMessage{
			Src: msg.Src,
			Id:  msg.Id,
			Payload: brachaWrapper{
				messageType: msg.Type,
				msg:         bm,
			},
			Paths: msg.Paths,
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
			Src:   msg.Src,
			Id:    msg.Id,
			Type:  bw.messageType,
			Paths: msg.Paths,
		})

		bm := bw.msg.(BrachaImprovedMessage)
		bdw.OriginalSrc = bm.Src
		bdw.OriginalId = bm.Id
		bdw.OriginalPayload = bm.Payload
		bdw.Included = bm.Included
	}

	bdw.Msgs = msgs
	return bdw
}
