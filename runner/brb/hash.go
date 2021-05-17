package brb

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/mitchellh/hashstructure/v2"
)

func MustHash(v interface{}) [sha256.Size]byte {
	return sha256.Sum256([]byte(fmt.Sprintf("%v", v)))
}

func MustHash2(v interface{}) [sha256.Size]byte {
	res, err := hashstructure.Hash(v, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}

	b := [sha256.Size]byte{}
	binary.LittleEndian.PutUint64(b[:], res)
	return b
}
