package process

import "encoding/binary"

const ControlIdMagic = 0x42

func IdToString(id uint64) string {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, id)
	return string(b)
}
