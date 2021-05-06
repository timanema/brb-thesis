package msg

type Message interface {
	Encode() ([]byte, error)
	Decode([]byte) error
}
