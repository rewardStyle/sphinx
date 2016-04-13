// Package sphinx is a Sphinx API library (not SphinxQL) that works with version 2.0.8 of Sphinx
package sphinx

import (
	"fmt"
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
}

// DefaultConfig provides sane defaults for the Sphinx Client
// - Listen on localhost with default Sphinx API port
// - Timeout of 10 seconds to connect to Sphinx server
// - Starting / Maximum connection pool size
var DefaultConfig = Config{
	Host:             "localhost",
	Port:             9312, // Default Sphinx API port
	ConnectTimeout:   time.Second * 10,
	StartingPoolSize: 1,
	MaxPoolSize:      30,
}

// SphinxClient represents a pooled connection to the Sphinx server
// Thread-safe after being opened.
type SphinxClient struct {
	config         Config
	ConnectionPool pool.Pool
}

// Limits:
// Offset: Distance from beginning of results
// Limit: Maximum matches to return
// Cutoff: Stop searching after this limit has been reached.
type Limits struct {
	Offset uint32
	Limit  uint32
	Cutoff uint32
}

type FieldWeight struct {
	FieldName   string
	FieldWeight uint32
}

// FilterValue is an approximation of struct st_filter type from sphinxclient.c
type FilterValue struct {
	Attribute string
	Type      Filter
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
}

// Init creates a SphinxClient with an initial connection pool to the Sphinx
// server.  We will need to
func (s *SphinxClient) Init(config *Config) error {

	if config == nil {
		config = &DefaultConfig
	}

	// Factory function that returns a new connection for use in the pool
	sphinxConnFactory := func() (net.Conn, error) {
		conn, err := net.DialTimeout(
			"tcp",
			fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
			s.config.ConnectTimeout,
		)
		if err != nil {
			return nil, err
		}
		return conn, rawInitializeSphinxConnection(conn)
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
func (s *SphinxClient) Query(q *SphinxQuery) error {
	conn, err := s.ConnectionPool.Get()
	if err != nil {
		// Type assertion as pool connection - have to since what is returned is
		// base interface type.
		if poolConn, ok := conn.(*pool.PoolConn); ok {
			poolConn.MarkUnusable()
		}
		return err
	}
	defer conn.Close()

	requestBuf, err := buildRequest(q)
	_ = requestBuf
	if err != nil {
		return err
	}

	// TODO: Send query and retrieve response

	return nil
}
