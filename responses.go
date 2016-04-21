package sphinx

import "bytes"

type responseHeader struct {
	status  uint16
	version uint16
	len     uint32
}

// Wrap bytes.Buffer so that I can define my own convenience methods on it
type ResponseReader struct {
	*bytes.Buffer
	internalErr error
}

// Reads header metadata from buffer
func (r *ResponseReader) ReadHeader() *responseHeader {
	// Read out each piece first since will be NOPs if there is a failure
	status := r.ReadWord()
	version := r.ReadWord()
	length := r.ReadInt()

	if r.internalErr != nil {
		return nil
	}

	return &responseHeader{
		status,
		version,
		length,
	}
}

func (r *ResponseReader) ReadWord() (word uint16) {
	if r.internalErr != nil {
		return
	}
	word = 0
	return
}

func (r *ResponseReader) ReadInt() (integer uint32) {
	if r.internalErr != nil {
		return
	}
	integer = 0
	return
}

// getResultsFromBuffer adds a
func getResultsFromBuffer(b *bytes.Buffer) ([]SphinxResult, error) {
	var reader = ResponseReader{Buffer: b, internalErr: nil}
	header := reader.ReadHeader()
	_ = header
	// Check at the end for an error, don't use anything before this
	return nil, nil
}
