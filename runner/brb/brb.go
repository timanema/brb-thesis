package brb

type Protocol interface {
	Receive(t uint8, src uint16, data []byte)
}
