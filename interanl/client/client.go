package client

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/tidwall/resp"
)

var (
	ErrOperationFailed = errors.New("operation failed")
)

type Client struct {
	addr     string
	connLock sync.Mutex
	conn     net.Conn
}

func New(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Client{
		addr: addr,
		conn: conn,
	}, nil
}

func (c *Client) Set(ctx context.Context, key string, value string) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	buf := &bytes.Buffer{}
	wr := resp.NewWriter(buf)
	err := wr.WriteArray([]resp.Value{resp.StringValue("SET"),
		resp.StringValue(key),
		resp.StringValue(value),
	})
	if err != nil {
		return nil
	}
	_, err = io.Copy(c.conn, buf)
	if err != nil {
		return err
	}
	var res bool
	err = binary.Read(c.conn, binary.BigEndian, &res)
	if err != nil {
		return err
	}
	if !res {
		return ErrOperationFailed
	}
	return nil
}
