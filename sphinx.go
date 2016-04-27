// Package sphinx is a Sphinx API library (not SphinxQL) that works with version 2.0.8 of Sphinx
package sphinx

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/fatih/pool"
)

// Config holds options needed for SphinxClient to connect to the server.
type Config struct {
	Host             string
	Port             int
	ConnectTimeout   time.Duration
	StartingPoolSize int
	MaxPoolSize      int
	// Query-specific config option
	MaxQueryTime time.Duration // Convert to milliseconds before sending
}

// NewDefaultConfig provides sane defaults for the Sphinx Client
// - Listen on localhost with default Sphinx API port
// - Timeout of 10 seconds to connect to Sphinx server
// - Starting / Maximum connection pool size
func NewDefaultConfig() *Config {
	return &Config{
		Host:             "0.0.0.0",
		Port:             9312, // Default Sphinx API port
		ConnectTimeout:   time.Second * 1,
		MaxQueryTime:     0,
		StartingPoolSize: 1,
		MaxPoolSize:      30,
	}
}

// SphinxClient represents a pooled connection to the Sphinx server
// Thread-safe after being opened.
type SphinxClient struct {
	Config         Config
	ConnectionPool pool.Pool
}

// Limits:
// Offset: Distance from beginning of results
// Limit: Maximum matches to return
// Cutoff: Stop searching after this limit has been reached.
// MaxMatches: Maximum # of matches to return (default 1000)
type Limits struct {
	Offset     uint32
	Limit      uint32
	Cutoff     uint32
	MaxMatches uint32
}

type FieldWeight struct {
	FieldName   string
	FieldWeight uint32
}

// FilterValue is an approximation of struct st_filter type from sphinxclient.c
type FilterValue struct {
	Attribute string
	Type      Filter
	Values    []uint64
	Min       uint64
	Max       uint64
	Fmin      float32
	Fmax      float32
	Exclude   uint32
}

type SphinxQuery struct {
	Keywords string
	Index    string
	// Discrete matching options
	MatchType MatchMode
	RankType  RankMode
	// Sorting options
	SortType     SortMode
	SortByString string
	// Filter options
	Filters []FilterValue
	// Offsets and max results
	QueryLimits  Limits
	FieldWeights []FieldWeight
	IndexWeights []FieldWeight
	// ID limits
	MinID uint64
	MaxID uint64

	MaxQueryTime time.Duration
	Comment      string
}

// DefaultQuery provides sane defaults for limits and index options.
// If value not specified, Go's zero value is the default.
func DefaultQuery() *SphinxQuery {
	return &SphinxQuery{
		Index: DefaultIndex,
		QueryLimits: Limits{
			Offset:     0,
			Limit:      20,
			Cutoff:     0,
			MaxMatches: 1000,
		},
	}
}

// Init creates a SphinxClient with an initial connection pool to the Sphinx
// server.  We will need to pass in a config or use the default.
func (s *SphinxClient) Init(config *Config) error {

	if config == nil {
		config = NewDefaultConfig()
	}

	s.Config = *config

	// Factory function that returns a new connection for use in the pool
	sphinxConnFactory := func() (net.Conn, error) {
		conn, err := net.DialTimeout(
			"tcp",
			fmt.Sprintf("%s:%d", s.Config.Host, s.Config.Port),
			s.Config.ConnectTimeout,
		)
		if err != nil {
			return nil, err
		}

		// Reset connect deadline to 0 after connection
		conn.SetDeadline(time.Now().Add(config.ConnectTimeout))
		log.Println("Initializing sphinx connection")
		err = rawInitializeSphinxConnection(conn)
		conn.SetDeadline(time.Time{})
		return conn, err
	}

	pool, err := pool.NewChannelPool(10, 30, sphinxConnFactory)

	s.ConnectionPool = pool

	return err
}

// Close closes the connection pool used by the client, which closes all
// outstanding connections
func (s *SphinxClient) Close() {
	s.ConnectionPool.Close()
}

// Query takes SphinxQuery objects and spawns off requests to Sphinx for them
// TODO: Decompose this into functions, remove debugging statements
func (s *SphinxClient) Query(q *SphinxQuery) (*SphinxResult, error) {
	// Build request first to avoid contention over connections in pool
	q.MaxQueryTime = s.Config.MaxQueryTime

	headerBuf, requestBuf, err := buildRequest(q)
	if err != nil {
		return nil, err
	}

	log.Println("Request for query built")

	conn, err := s.ConnectionPool.Get()
	if err != nil {
		// Type assertion as pool connection - have to since what is returned is
		// base interface type.
		if poolConn, ok := conn.(*pool.PoolConn); ok {
			poolConn.MarkUnusable()
		}
		return nil, err
	}
	defer conn.Close()

	_, err = headerBuf.WriteTo(conn)
	if err != nil {
		return nil, err
	}
	_, err = requestBuf.WriteTo(conn)
	if err != nil {
		return nil, err
	}
	log.Println("Wrote request to server")

	responseHeader, err := readHeader(conn)
	if err != nil {
		return nil, err
	}

	// Now need to read the remainder of the response into the buffer
	// FIXME: Check len to make sure reasonable
	responseBytes := make([]byte, responseHeader.len)
	_, err = io.ReadFull(conn, responseBytes)
	if err != nil {
		return nil, err
	}

	result, err := getResultFromBuffer(responseHeader, bytes.NewBuffer(responseBytes))

	return result, err
}

// NewSearch gives a better API to creating a query object that can be passed to
// Query.
func NewSearch(keywords, index, comment string) *SphinxQuery {
	q := DefaultQuery()
	q.Keywords = keywords
	if index == "" {
		q.Index = DefaultIndex
	} else {
		q.Index = index
	}

	q.Comment = comment
	return q
}
