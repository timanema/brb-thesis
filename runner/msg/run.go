package msg

const TriggerMessageType uint8 = 5
const WrapperDataType uint8 = 6
const MessageDeliveredType uint8 = 7

type TriggerMessage struct {
	Id      uint32
	Payload interface{}
}

type WrapperDataMessage struct {
	T    uint8
	Id   uint32
	Data interface{}
}

type MessageDelivered struct {
	Id      uint32
	Payload interface{}
}
