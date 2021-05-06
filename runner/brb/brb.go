package brb

type Protocol interface {
	// uid is used for tracking the message throughout the network (for statistics)
	Receive(messageType uint8, src uint16, uid uint32, data []byte)
}
