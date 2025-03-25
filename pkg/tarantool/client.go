package tarantool

import (
	"context"
	"github.com/pkg/errors"
	"github.com/tarantool/go-tarantool"
)

// Client manages Tarantool connections
type Client struct {
	conn *tarantool.Connection
}

// NewClient creates new Tarantool client
func NewClient(addr string, opts tarantool.Opts) (*Client, error) {
	conn, err := tarantool.Connect(addr, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to Tarantool")
	}

	return &Client{conn: conn}, nil
}

// HealthCheck verifies connection using Ping
func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.conn.Ping()
	return errors.Wrap(err, "connection check failed")
}

// Close releases resources
func (c *Client) Close() error {
	return c.conn.Close()
}
