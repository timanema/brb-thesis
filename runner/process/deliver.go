package process

// Used as abstraction for BRB protocols

type Application interface {
	// uid is used for tracking the message throughout the network (for statistics)
	Deliver(uid uint32, payload []byte)
}

type Network interface {
	// uid is used for tracking the message throughout the network (for statistics)
	Send(messageType uint8, dest uint16, uid uint32, data []byte)
}
