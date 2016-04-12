// Package sphinx is a Sphinx API library (not SphinxQL) that works with version 2.0.8 of Sphinx
package sphinx

import (
	"bytes"
	"fmt"
	"net"
)

// Config holds options needed for SphinxClient to connect to the server.
type Config struct {
	Host string
	Port int
}

// SphinxClient represents a pooled connection to the Sphinx server
// Thread-safe after being opened.
type SphinxClient struct {
	config     Config
	Connection net.Conn
}

// Open creates a connection to the Sphinx server
func (s *SphinxClient) Open(config *Config) error {
	// Already connected
	if s.Connection != nil {
		return nil
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.config.Host, s.config.Port))
	if err != nil {
		return err
	}

	s.Connection = conn

	reqBuffer := bytes.NewBuffer(make([]byte, 0, 16))
	// Should this be a bytes.Buffer type?
	addWordToBuffer(reqBuffer, SEARCHD_COMMAND_PERSIST)
	addWordToBuffer(reqBuffer, 0)
	addIntToBuffer(reqBuffer, 4)
	addIntToBuffer(reqBuffer, 1)

	_, err = reqBuffer.WriteTo(conn)

	return err
}

// Close closes the socket used by the client
func (s *SphinxClient) Close() error {
	return s.Connection.Close()
}
