package process

// Used as abstraction for BRB protocols

type Application interface {
	Deliver(payload []byte)
}

type Network interface {
	Send(t uint8, dest uint16, data []byte)
}
