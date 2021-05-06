package process

import "encoding/binary"

const ControlIdMagic = 0x42

func IdToString(id uint16) string {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, id)
	return string(b)
}
