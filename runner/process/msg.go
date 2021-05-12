package process

type Message struct {
	Src  uint64
	Ctl  bool
	Type uint8
	Data interface{}
}
