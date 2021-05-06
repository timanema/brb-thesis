package brb

// Used as abstraction for BRB protocols
// uid is used for tracking the message throughout the network (for statistics)

type Application interface {
	Deliver(uid uint32, payload []byte)
}

type Network interface {
	Send(messageType uint8, dest uint16, uid uint32, data []byte)
}

type Protocol interface {
	Receive(messageType uint8, src uint16, uid uint32, data []byte)

	Send(uid uint32, payload []byte)
}
