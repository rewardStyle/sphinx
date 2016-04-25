package sphinx

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// SafeWriter is used to consolidate the errors found when writing to a buffer
// on error, subsequent writes are NOPs.  Can't satisfy io.Writer since don't
// want to have same semantics (retains buffer).
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

// AddStringToBuffer adds a string preceded by length (as uint32) to the buffer
func (s *SafeWriter) AddStringToBuffer(inputString string) {
	// Write length of string first (before checking if s.err != nil) so that
	// don't clobber it if it is.
	s.AddIntToBuffer(uint32(len(inputString)))
	if s.err != nil {
		return
	}
	_, s.err = s.internalBuf.WriteString(inputString)
}

// AddIntToBuffer writes an unsigned int (32 bits) in network byte order
func (s *SafeWriter) AddIntToBuffer(unsigned uint32) {
	if s.err != nil {
		return
	}

	s.err = binary.Write(s.internalBuf, binary.BigEndian, unsigned)
}

// AddFloatToBuffer writes 32-bit floating point value to the buffer
func (s *SafeWriter) AddFloatToBuffer(f float32) {
	if s.err != nil {
		return
	}
	s.err = binary.Write(s.internalBuf, binary.BigEndian, f)
}

// AddInt64ToBuffer writes an unsigned int (64 bits) in network byte order
func (s *SafeWriter) AddInt64ToBuffer(unsigned uint64) {
	if s.err != nil {
		return
	}

	s.err = binary.Write(s.internalBuf, binary.BigEndian, unsigned)
}

// AddFilterToBuffer is a helper function that writes one of the query filters to
// the internal buffer so that it can be sent to the server.
func (s *SafeWriter) AddFilterToBuffer(filter *FilterValue) {
	if s.err != nil {
		return
	}
	s.AddStringToBuffer(filter.Attribute)
	s.AddIntToBuffer(uint32(filter.Type))

	switch filter.Type {
	case SPH_FILTER_VALUES:
		s.AddIntToBuffer(uint32(len(filter.Values)))
		for _, value := range filter.Values {
			s.AddInt64ToBuffer(value)
		}
	case SPH_FILTER_RANGE:
		s.AddInt64ToBuffer(filter.Min)
		s.AddInt64ToBuffer(filter.Max)
	case SPH_FILTER_FLOATRANGE:
		s.AddFloatToBuffer(filter.Fmin)
		s.AddFloatToBuffer(filter.Fmax)
	}

	s.AddIntToBuffer(filter.Exclude)
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

	if size != 0 {
		swriter = &SafeWriter{
			internalBuf: bytes.NewBuffer(make([]byte, 0, int(size))),
			err:         nil,
		}
		return
	}
	swriter = &SafeWriter{
		internalBuf: new(bytes.Buffer),
		err:         nil,
	}
	return
}

// calculateRequestLength returns the length needed for the request body as
// calculated by sphinxclient.c
// (See line 1825 in sphinx_run_queries)
func calculateRequestLength(buf *bytes.Buffer) uint32 {
	return uint32(buf.Len() + 8)
}

// Mostly taken from sphinx_add_query
func buildInternalQuery(query *SphinxQuery) (buf *bytes.Buffer, err error) {
	var internalQuery = NewSafeWriter(0)
	internalQuery.AddIntToBuffer(query.QueryLimits.Offset)
	internalQuery.AddIntToBuffer(query.QueryLimits.Limit)
	internalQuery.AddIntToBuffer(uint32(query.MatchType)) // Match Mode
	internalQuery.AddIntToBuffer(uint32(query.RankType))  // Rank Mode
	internalQuery.AddIntToBuffer(uint32(query.SortType))
	internalQuery.AddStringToBuffer(query.SortByString)
	internalQuery.AddStringToBuffer(query.Keywords)
	internalQuery.AddIntToBuffer(0) // this is client->num_weights, which AFAICT is always zero
	// Skip the weight values, since will never exist.
	internalQuery.AddStringToBuffer(query.Index)
	internalQuery.AddIntToBuffer(1)
	internalQuery.AddInt64ToBuffer(query.MinID)
	internalQuery.AddInt64ToBuffer(query.MaxID)
	internalQuery.AddIntToBuffer(uint32(len(query.Filters)))

	for i := range query.Filters {
		internalQuery.AddFilterToBuffer(&query.Filters[i])
	}

	// Query limits and group by options (don't support setting)
	internalQuery.AddIntToBuffer(SPH_GROUPBY_FUNC_DEFAULT)
	internalQuery.AddStringToBuffer(SPH_GROUPBY_DEFAULT)
	internalQuery.AddIntToBuffer(query.QueryLimits.MaxMatches)
	internalQuery.AddStringToBuffer(SPH_GROUPBY_SORT_DEFAULT)
	internalQuery.AddIntToBuffer(query.QueryLimits.Cutoff)
	// Retry stuff
	internalQuery.AddIntToBuffer(0) // Retry count
	internalQuery.AddIntToBuffer(0) // Retry delay

	internalQuery.AddStringToBuffer("") // Group distinct filler
	internalQuery.AddIntToBuffer(0)     // Geoanchor filler

	// Add index weights
	internalQuery.AddIntToBuffer(uint32(len(query.IndexWeights)))
	for _, weight := range query.IndexWeights {
		internalQuery.AddStringToBuffer(weight.FieldName)
		internalQuery.AddIntToBuffer(weight.FieldWeight)
	}

	// Convert to representation in milliseconds
	internalQuery.AddIntToBuffer(
		uint32(query.MaxQueryTime / time.Millisecond),
	)

	// Add field weights
	internalQuery.AddIntToBuffer(uint32(len(query.FieldWeights)))
	for _, weight := range query.FieldWeights {
		internalQuery.AddStringToBuffer(weight.FieldName)
		internalQuery.AddIntToBuffer(weight.FieldWeight)
	}

	internalQuery.AddStringToBuffer(query.Comment)

	// Overrides - unsupported
	internalQuery.AddIntToBuffer(0)

	// Select list - unsupported
	internalQuery.AddStringToBuffer("")

	return internalQuery.Buf(), internalQuery.Err()
}

func rawInitializeSphinxConnection(sphinxConnection net.Conn) (err error) {
	// Have to send protocol major version, make sure agrees (see net_connect_get)
	err = binary.Write(sphinxConnection, binary.BigEndian, uint32(MAJOR_PROTOCOL_VERSION))
	if err != nil {
		return
	}
	log.Println("Sending major protocol version")
	// Now get protocol version back and compare.
	versionBytes := make([]byte, 4)
	_, err = io.ReadFull(sphinxConnection, versionBytes)
	if err != nil {
		return
	}
	serverVersion := binary.BigEndian.Uint32(versionBytes)
	if serverVersion < MAJOR_PROTOCOL_VERSION {
		err = fmt.Errorf(
			"Expected version to be >= `%v` but got `%v`\n",
			MAJOR_PROTOCOL_VERSION, serverVersion,
		)
	}

	log.Println("Sending header to server for SEARCHD_COMMAND_PERSIST")

	// Send header to establish connection
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
	log.Println("Wrote header to sphinx server.")
	return
}

func buildRequest(query *SphinxQuery) (headerBuf *bytes.Buffer, requestBuf *bytes.Buffer, err error) {
	requestBuf, err = buildInternalQuery(query)
	if err != nil {
		return
	}
	requestLength := calculateRequestLength(requestBuf)
	headerBuffer := NewSafeWriter(16)

	// Build query header based on request length - see sphinx_run_queries
	headerBuffer.AddWordToBuffer(uint16(SEARCHD_COMMAND_SEARCH))
	headerBuffer.AddWordToBuffer(VER_COMMAND_SEARCH)
	headerBuffer.AddIntToBuffer(requestLength)
	headerBuffer.AddIntToBuffer(0) // Dummy body?
	headerBuffer.AddIntToBuffer(1) // Number of requests

	headerBuf = headerBuffer.Buf()
	err = headerBuffer.Err()

	return
}
