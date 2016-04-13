package sphinx

import (
	"bytes"
	"encoding/binary"
	"net"
)

// SafeWriter is used to consolidate the errors found when writing to a buffer
// on error, subsequent writes are NOPs.  Can't satisfy io.Writer since don't
// want to have same semantics.
type SafeWriter struct {
	internalBuf *bytes.Buffer
	err         error
}

// AddWordToBuffer writes an unsigned short (16 bits) in network byte order
func (s *SafeWriter) AddWordToBuffer(short uint16) {
	if s.err != nil {
		return
	}
	s.err = binary.Write(s.internalBuf, binary.BigEndian, short)
}

// AddIntToBuffer writes an unsigned int (32 bits) in network byte order
func (s *SafeWriter) AddIntToBuffer(unsigned uint32) {
	if s.err != nil {
		return
	}

	s.err = binary.Write(s.internalBuf, binary.BigEndian, unsigned)
}

// Write sends bytes to backing buffer (or if err already set, a NOP)
func (s *SafeWriter) Write(data []byte) {
	if s.err != nil {
		return
	}
	_, s.err = s.internalBuf.Write(data)
}

func (s *SafeWriter) Err() error {
	return s.err
}

func (s *SafeWriter) Buf() *bytes.Buffer {
	return s.internalBuf
}

func NewSafeWriter(size uint) (swriter *SafeWriter) {

	if size == 0 {
		swriter = &SafeWriter{
			internalBuf: bytes.NewBuffer(make([]byte, 0, int(size))),
			err:         nil,
		}
		return
	}
	swriter = &SafeWriter{}
	return
}

// calculateRequestLength returns the length needed for the request body as
// calculated by sphinxclient.c
// (See line 1825 in sphinx_run_queries)
func calculateRequestLength(buf *bytes.Buffer) uint32 {
	return uint32(buf.Len() + 8)
}

func buildInternalQuery(query SphinxQuery) *bytes.Buffer {
	// TODO: Implement
	return nil
}

func rawInitializeSphinxConnection(sphinxConnection net.Conn) (err error) {
	reqBuffer := NewSafeWriter(16)
	// Should this be a bytes.Buffer type?
	reqBuffer.AddWordToBuffer(uint16(SEARCHD_COMMAND_PERSIST)) // Command
	reqBuffer.AddWordToBuffer(0)                               // Dummy Version
	reqBuffer.AddIntToBuffer(4)                                // Dummy body length
	reqBuffer.AddIntToBuffer(1)                                // Dummy body

	err = reqBuffer.Err()
	if err != nil {
		return
	}

	_, err = reqBuffer.Buf().WriteTo(sphinxConnection)
	return
}

func buildRequest(query SphinxQuery) (buf *bytes.Buffer, err error) {
	// TODO: Calculate request length as in calc_req_len in sphinxclient.c
	internalQuery := buildInternalQuery(query)
	requestLength := calculateRequestLength(internalQuery)
	headerBuffer := NewSafeWriter(16)

	headerBuffer.AddWordToBuffer(uint16(SEARCHD_COMMAND_SEARCH))
	headerBuffer.AddWordToBuffer(VER_COMMAND_SEARCH)
	headerBuffer.AddIntToBuffer(requestLength)
	headerBuffer.AddIntToBuffer(0) // Dummy body?
	headerBuffer.AddIntToBuffer(1) // Number of requests

	buf = headerBuffer.Buf()
	err = headerBuffer.Err()

	return
}
