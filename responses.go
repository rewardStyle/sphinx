package sphinx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"time"
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
	// FIXME: Get rid of internalErr and just use panic on temporary error - will
	// be handled by deferred error handler.
	internalErr error
}

type ResponseAttribute struct {
	Name string
	Type uint32
	// In C version, Attr is one union value discriminated by AttrType
	AttrInt32Value  uint32
	AttrInt64Value  uint64
	AttrFloatValue  float32
	AttrStringValue string
	AttrMultiValue  []uint64
}

type Match struct {
	DocId  uint64 // Have to cast if id64 != 1
	Weight uint32
	Attrs  []ResponseAttribute
}

type Word struct {
	Word string
	Docs uint32
	Hits uint32
}

// SphinxResult is a container for all of the fields returned by Sphinx
type SphinxResult struct {
	Fields      []string
	Matches     []Match
	Total       uint32
	TotalFound  uint32
	SearchTime  time.Duration // Returned in milliseconds
	Words       []Word
	MakeID64Bit uint32
	// Attribute names and types will be the same for all of the matches
	AttrNames []string
	AttrTypes []uint32
}

// Read response header
func readHeader(r io.Reader) (header *ResponseHeader, err error) {
	const headerSize = 8
	headerBytes := make([]byte, headerSize)

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
		panic(r.internalErr)
	}
	wordBytes := r.Next(2)
	if len(wordBytes) != 2 {
		r.internalErr = fmt.Errorf(
			"Expected to read 2 bytes for ReadWord, got %v\n",
			len(wordBytes))
		panic(r.internalErr)
	}

	word = binary.BigEndian.Uint16(wordBytes)
	return
}

// ReadInt parses a 32 bit integer in BigEndian byte order
func (r *ResponseReader) ReadInt() (integer uint32) {
	if r.internalErr != nil {
		panic(r.internalErr)
	}
	intBytes := r.Next(4)
	if len(intBytes) != 4 {
		r.internalErr = fmt.Errorf(
			"Expected to read 4 bytes for ReadInt, got %v\n",
			len(intBytes),
		)
		panic(r.internalErr)
	}

	integer = binary.BigEndian.Uint32(intBytes)
	return
}

// ReadFloat32 parses a 32 bit floating point value in BigEndian byte order
func (r *ResponseReader) ReadFloat32() (float float32) {
	r.internalErr = binary.Read(r, binary.BigEndian, &float)
	if r.internalErr != nil {
		panic(r.internalErr)
	}
	return
}

// Skip safely skips the next n bytes in the buffer.
func (r *ResponseReader) Skip(n int) {
	if r.internalErr != nil {
		panic(r.internalErr)
	}

	temp := r.Next(n)
	if len(temp) != n {
		r.internalErr = fmt.Errorf(
			"Expected to skip %v bytes, got %v\n",
			n, len(temp),
		)
		panic(r.internalErr)
	}
}

// Read64BitInt reads 8 bytes in BigEndian format and parses it as an int64
func (r *ResponseReader) Read64BitInt() uint64 {
	if r.internalErr != nil {
		panic(r.internalErr)
	}

	temp := r.Next(8)
	if len(temp) != 8 {
		r.internalErr = fmt.Errorf(
			"Expected to read %v bytes for 64-bit int, got %v\n",
			8, len(temp),
		)
		panic(r.internalErr)
	}

	return binary.BigEndian.Uint64(temp)
}

// ReadMatches reads out sequence of Match structs from response buffer
func (r *ResponseReader) ReadMatches(
	n int, make64bit bool, numAttrs int, attrTypes []uint32,
) (matches []Match) {
	if n < 0 {
		panic(fmt.Errorf("Invalid match number for response: %v\n", n))
	}
	if n == 0 {
		return
	}

	matches = make([]Match, n)
	for matchNum := range matches {
		var (
			DocID uint64
		)

		if make64bit {
			DocID = r.Read64BitInt()
		} else {
			DocID = uint64(r.ReadInt())
		}

		matches[matchNum].DocId = DocID
		matches[matchNum].Weight = r.ReadInt()
		matches[matchNum].Attrs = make([]ResponseAttribute, numAttrs)

		// Unpack attribute and save to Match object
		for j := 0; j < numAttrs; j++ {
			attributeType := attrTypes[j]

			matches[matchNum].Attrs[j].Type = attributeType

			switch attributeType {
			// See sphinx_run_queries... this bit is pretty schwifty.  We only read 4
			// bytes (high-order) for each attr value, but need to skip 4 for 64-bit integer
			// size.  This seems like Aksyonoff is working around ints having different
			// sizes on different clients.
			case SPH_ATTR_MULTI, SPH_ATTR_MULTI64:
				// TODO: Sanity checking on numberValues
				numberValues := int(r.ReadInt())
				// If SPH_ATTR_MULTI64, has half as many words as would expect in same
				// sized response.
				if attributeType == SPH_ATTR_MULTI64 {
					numberValues /= 2
					matches[matchNum].Attrs[j].AttrMultiValue = make([]uint64, numberValues)
					for i := 0; i < numberValues; i++ {
						matches[matchNum].Attrs[j].AttrMultiValue[i] = uint64(r.ReadInt())
						r.Skip(4)
					}
				} else {
					matches[matchNum].Attrs[j].AttrMultiValue = make([]uint64, numberValues)
					for i := 0; i < numberValues; i++ {
						matches[matchNum].Attrs[j].AttrMultiValue[i] = uint64(r.ReadInt())
					}
				}
			case SPH_ATTR_FLOAT:
				matches[matchNum].Attrs[j].AttrFloatValue = r.ReadFloat32()
			case SPH_ATTR_BIGINT:
				matches[matchNum].Attrs[j].AttrInt64Value = r.Read64BitInt()
			case SPH_ATTR_STRING:
				matches[matchNum].Attrs[j].AttrStringValue = r.ReadString()
			default:
				// YOLO it's an int
				matches[matchNum].Attrs[j].AttrInt32Value = r.ReadInt()
			}
		}

	}
	return
}

// ReadWords reads out word, document and hit count for all words in response
// buffer.
func (r *ResponseReader) ReadWords(n int) (words []Word) {
	if n < 0 {
		panic(fmt.Errorf("Invalid word count for response: %v\n", n))
	}
	if n == 0 {
		return
	}

	words = make([]Word, n)
	for wordIndex := range words {
		words[wordIndex].Word = r.ReadString()
		words[wordIndex].Docs = r.ReadInt()
		words[wordIndex].Hits = r.ReadInt()
	}

	return
}

// ReadString parses a string out (with length as uint32 preceding it) from the
// buffer.
func (r *ResponseReader) ReadString() (s string) {
	stringLength := r.ReadInt()

	// Don't ovewrite error if already set
	if stringLength < 0 {
		if r.internalErr == nil {
			r.internalErr = fmt.Errorf("Invalid string length: %v\n", stringLength)
			panic(r.internalErr)
		}
	}

	// Valid length - if no string found
	if stringLength == 0 {
		return
	}

	if r.internalErr != nil {
		panic(r.internalErr)
	}

	stringBytes := r.Next(int(stringLength))
	if len(stringBytes) != int(stringLength) {
		r.internalErr = fmt.Errorf(
			"Expected to read %v bytes for ReadString, got %v\n",
			stringLength,
			len(stringBytes),
		)
		panic(r.internalErr)
	}

	s = string(stringBytes)
	return
}

// parseResponseBody gets specifically the response object from the buffer, after
// the header has already been read.
func parseResponseBody(r ResponseReader) (result *SphinxResult, searchError error) {
	// Allows better error handling, since don't have to check all intermediate
	// steps for errors.
	defer func() {
		if r := recover(); r != nil {
			result = nil
			searchError = fmt.Errorf("%v", r)
		}
		return
	}()
	status := r.ReadInt()

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

	// FIXME: Sanity check numFields
	result.Fields = make([]string, int(numFields))
	for i := 0; i < int(numFields); i++ {
		result.Fields[i] = r.ReadString()
	}

	// FIXME: Sanity check numAttrs
	numAttrs := int(r.ReadInt())

	result.AttrNames = make([]string, numAttrs)
	result.AttrTypes = make([]uint32, numAttrs)

	// Attribute types, names, and count will be the same for all
	// Build factory function to create once have read schema.
	for i := 0; i < numAttrs; i++ {
		result.AttrNames[i] = r.ReadString()
		result.AttrTypes[i] = r.ReadInt()
	}

	// FIXME: Sanity check numMatches
	numMatches := int(r.ReadInt())
	result.Matches = make([]Match, numMatches)

	result.MakeID64Bit = r.ReadInt()
	var make64bit bool
	if result.MakeID64Bit == 1 {
		make64bit = true
	} else {
		make64bit = false
	}
	result.Matches = r.ReadMatches(numMatches, make64bit, numAttrs, result.AttrTypes)

	result.Total = r.ReadInt()
	result.TotalFound = r.ReadInt()
	result.SearchTime = time.Duration(r.ReadInt()) * time.Millisecond

	numWords := int(r.ReadInt()) // FIXME: Sanity check numWords

	result.Words = r.ReadWords(numWords)

	return
}

// getResultsFromBuffer parses out the response data from the buffer
// and make sure that everything is okay with the response.  Mainly based
// on net_get_response and latter part of sphinx_run_queries
func getResultFromBuffer(header *ResponseHeader, b *bytes.Buffer) (result *SphinxResult, searchError error) {
	defer func() {
		if r := recover(); r != nil {
			result = nil
			searchError = fmt.Errorf("%v", r)
		}
		return
	}()
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
