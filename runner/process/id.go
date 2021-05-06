package process

import "encoding/binary"

func idToString(id uint16) string {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, id)
	return string(b)
}

func idFromString(id string) uint16 {
	return binary.BigEndian.Uint16([]byte(id))
}
