package sphinx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
)

// ResponseHeader is the beginning of each response from Sphinx that gives
// metadata to the version, status, and length of the response.
type ResponseHeader struct {
	status  uint16
	version uint16
	len     uint32
}

// ResponseReader wraps bytes.Buffer so that I can define my own convenience
// methods on it.  It works like SafeWriter in terms of error handling.
type ResponseReader struct {
	*bytes.Buffer
	internalErr error
}

// SphinxResult is a container for all of the fields and
type SphinxResult struct{}

// Read response header - gives header size
func readHeader(r io.Reader) (header *ResponseHeader, err error) {
	const headerSize = 8
	headerBytes := make([]byte, 8)

	_, err = io.ReadFull(r, headerBytes)
	if err != nil {
		return
	}

	status := binary.BigEndian.Uint16(headerBytes[0:2])
	version := binary.BigEndian.Uint16(headerBytes[2:4])
	len := binary.BigEndian.Uint32(headerBytes[4:8])

	header = &ResponseHeader{
		status,
		version,
		len,
	}
	return
}

// ReadWord parses 16 bit integer (short) in BigEndian byte order
func (r *ResponseReader) ReadWord() (word uint16) {
	if r.internalErr != nil {
		return
	}
	wordBytes := r.Next(2)
	if len(wordBytes) != 2 {
		r.internalErr = fmt.Errorf(
			"Expected to read 2 bytes for ReadWord, got %v\n",
			len(wordBytes))
		return
	}

	word = binary.BigEndian.Uint16(wordBytes)
	return
}

// ReadInt parses a 32 bit integer in BigEndian byte order
func (r *ResponseReader) ReadInt() (integer uint32) {
	if r.internalErr != nil {
		return
	}
	intBytes := r.Next(4)
	if len(intBytes) != 4 {
		r.internalErr = fmt.Errorf(
			"Expected to read 4 bytes for ReadInt, got %v\n",
			len(intBytes),
		)
	}

	return
}

// ReadString parses a string out (with length as uint32 preceding it) from the
// buffer.
func (r *ResponseReader) ReadString() (s string) {
	stringLength := r.ReadInt()

	// This is a bit tortured - return if already have error or if string length
	// is invalid, but shouldn't overwrite error if already there.
	if stringLength < 0 {
		if r.internalErr == nil {
			r.internalErr = fmt.Errorf("Invalid string length: %v\n", stringLength)
			return
		}
	}

	if stringLength == 0 {
		return
	}

	if r.internalErr != nil {
		return
	}

	stringBytes := r.Next(int(stringLength))
	if len(stringBytes) != int(stringLength) {
		r.internalErr = fmt.Errorf(
			"Expected to read %v bytes for ReadString, got %v\n",
			stringLength,
			len(stringBytes),
		)
	}

	s = string(stringBytes)
	return
}

// parseResponseBody gets specifically the response object from the buffer, after
// the header has already been read.
func parseResponseBody(r ResponseReader) (result *SphinxResult, searchError error) {
	status := r.ReadInt()
	if r.internalErr != nil {
		return nil, r.internalErr
	}

	// Response has its own status
	switch status {
	case SEARCHD_OK:
		break
	case SEARCHD_WARNING:
		warning := r.ReadString()
		log.Printf("Warning: %v\n", warning)
	case SEARCHD_ERROR:
		errMsg := r.ReadString()
		searchError = errors.New(errMsg)
		return
	}

	result = new(SphinxResult)

	numFields := r.ReadInt()
	if r.internalErr != nil {
		return nil, r.internalErr
	}

	fields := make([]string, int(numFields))
	for i := 0; i < int(numFields); i++ {
		fields[i] = r.ReadString()
	}

	numAttrs := r.ReadInt()
	if r.internalErr != nil {
		return nil, r.internalErr
	}

	for i := 0; i < int(numAttrs); i++ {
		_ = i
	}

	return nil, r.internalErr
}

// getResultsFromBuffer parses out the response data from the buffer
// and make sure that everything is okay with the response.  Mainly based
// on net_get_response and latter part of sphinx_run_queries
func getResultFromBuffer(header *ResponseHeader, b *bytes.Buffer) (result *SphinxResult, searchError error) {
	var reader = ResponseReader{Buffer: b, internalErr: nil}

	switch header.status {
	case SEARCHD_OK:
		fallthrough
	case SEARCHD_WARNING:
		warning := reader.ReadString()
		log.Printf("Warning: %v\n", warning)
	case SEARCHD_ERROR:
		fallthrough
	case SEARCHD_RETRY:
		searchError = errors.New(reader.ReadString())
	default:
		searchError = fmt.Errorf("Unknown status code %v from response\n", header.status)
	}

	// Now need to parse out responses as in sphinx_run_queries
	// and return them if needed.  Know only have 1 result (if any), since we
	// always send one query-at-a-time.

	result, searchError = parseResponseBody(reader)
	return

}
