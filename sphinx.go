// Package sphinx is a Sphinx API library (not SphinxQL) that works with version 2.0.8 of Sphinx
package sphinx

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/fatih/pool"
)

// Config holds options needed for SphinxClient to connect to the server.
type Config struct {
	Host           string
	Port           int
	ConnectTimeout time.Duration
	PoolSize       int
}

// SphinxClient represents a pooled connection to the Sphinx server
// Thread-safe after being opened.
type SphinxClient struct {
	config         Config
	ConnectionPool pool.Pool
}

type SphinxQuery struct {
}

func rawInitializeSphinxConnection(sphinxConnection net.Conn) error {
	reqBuffer := bytes.NewBuffer(make([]byte, 0, 16))
	// Should this be a bytes.Buffer type?
	addWordToBuffer(reqBuffer, SEARCHD_COMMAND_PERSIST)
	addWordToBuffer(reqBuffer, 0) // Version
	addIntToBuffer(reqBuffer, 4)  // Dummy body length
	addIntToBuffer(reqBuffer, 1)  // Dummy body

	_, err := reqBuffer.WriteTo(sphinxConnection)
	return err
}

// Init creates a SphinxClient with an initial connection pool to the Sphinx
// server.  We will need to
func (s *SphinxClient) Init(config *Config) error {

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

// Send command to Sphinx, one of which can be a query
func (s *SphinxClient) doRequest(command, version int, request []byte) error {
	return nil
}

// Close closes the socket used by the client
func (s *SphinxClient) Close() {
	s.ConnectionPool.Close()
}

// Query takes SphinxQuery objects and spawns off requests to Sphinx for them
func (s *SphinxClient) Query(q SphinxQuery) error {
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

	// TODO: Send query and retrieve response

	return nil
}
