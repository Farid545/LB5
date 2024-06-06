package datastore

import (
	"encoding/binary"
	"fmt"
)

type entry struct {
	key, value string
}

// Encode converts the entry to a byte slice
func (e *entry) Encode() []byte {
	fmt.Println("Encoding entry:", e.key, e.value)
	kl := len(e.key) // Key length
	vl := len(e.value) // Value length
	size := kl + vl + 12
	res := make([]byte, size)
	binary.LittleEndian.PutUint32(res, uint32(size))
	binary.LittleEndian.PutUint32(res[4:], uint32(kl))
	copy(res[8:], e.key)
	binary.LittleEndian.PutUint32(res[kl+8:], uint32(vl))
	copy(res[kl+12:], e.value)
	return res
}

// Decode converts a byte slice back to an entry
func (e *entry) Decode(input []byte) {
	fmt.Println("Decoding entry")
	kl := binary.LittleEndian.Uint32(input[4:])
	keyBuf := make([]byte, kl)
	copy(keyBuf, input[8:kl+8])
	e.key = string(keyBuf)

	vl := binary.LittleEndian.Uint32(input[kl+8:]) // Get value length
	valBuf := make([]byte, vl)
	copy(valBuf, input[kl+12:kl+12+vl])
	e.value = string(valBuf)
}
