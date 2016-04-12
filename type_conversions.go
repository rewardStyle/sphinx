package sphinx

import (
	"bytes"
	"encoding/binary"
)

func addWordToBuffer(b *bytes.Buffer, short uint16) (err error) {
	// All this is to write an unsigned short (16 bytes) in network byte order
	buf := make([]byte, 0, 2)
	binary.BigEndian.PutUint16(buf, short)
	_, err = b.Write(buf)
	return
}

// addIntToBuffer puts a 32 bit integer in network byte order to buffer
func addIntToBuffer(b *bytes.Buffer, unsigned uint32) (err error) {
	buf := make([]byte, 0, 4)
	binary.BigEndian.PutUint32(buf, unsigned)
	_, err = b.Write(buf)
	return
}
