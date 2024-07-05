package client

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/tidwall/resp"
)

var (
	ErrOperationFailed = errors.New("operation failed")
	ErrTimeIsOut       = errors.New("time is out")
)

type Client struct {
	addr     string
	connLock sync.Mutex
	conn     net.Conn
}
type GetResult struct {
	value string
	err   error
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
		return err
	}
	_, err = io.Copy(c.conn, buf)
	if err != nil {
		return err
	}
	ch := make(chan error)
	go func() {
		var res bool
		err = binary.Read(c.conn, binary.BigEndian, &res)
		if err != nil {
			ch <- err
		}
		if !res {
			ch <- ErrOperationFailed
		}
		ch <- nil

	}()
	select {
	case <-ctx.Done():
		return ErrTimeIsOut
	case err := <-ch:
		if err != nil {
			return ErrOperationFailed
		}
		return nil
	}
}
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	buf := &bytes.Buffer{}
	wr := resp.NewWriter(buf)
	err := wr.WriteArray([]resp.Value{resp.StringValue("GET"),
		resp.StringValue(key)})
	if err != nil {
		return "", err
	}
	_, err = io.Copy(c.conn, buf)
	if err != nil {
		return "", err
	}
	ch := make(chan GetResult)
	go func() {
		res, err := c.readResult()
		ch <- GetResult{
			value: res,
			err:   err,
		}
	}()
	select {
	case <-ctx.Done():
		return "", ErrTimeIsOut
	case resp := <-ch:
		return resp.value, resp.err
	}
}
func (c *Client) Hello(ctx context.Context, m map[string]string) error {
	mapString := writeMapResp(m)
	buf := &bytes.Buffer{}
	wr := resp.NewWriter(buf)
	err := wr.WriteArray([]resp.Value{resp.StringValue("HELLO"),
		resp.StringValue(mapString)})
	if err != nil {
		return err
	}
	_, err = io.Copy(c.conn, buf)
	if err != nil {
		return err
	}
	return nil
}
func (c *Client) readResult() (string, error) {
	var res bool
	err := binary.Read(c.conn, binary.BigEndian, &res)
	if err != nil {
		return "", err
	}
	if !res {
		return "", ErrOperationFailed
	}
	var size int64
	if err := binary.Read(c.conn, binary.BigEndian, &size); err != nil {
		return "", err
	}
	valueBuf := make([]byte, size)
	if _, err := c.conn.Read(valueBuf); err != nil {
		return "", nil
	}
	return string(valueBuf), nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func writeMapResp(m map[string]string) string {
	buf := bytes.Buffer{}
	buf.WriteString("%" + fmt.Sprintf("%v\r\n", len(m)))
	for key, val := range m {
		buf.WriteString(fmt.Sprintf("+%s\r\n", key))
		buf.WriteString(fmt.Sprintf(":%s\r\n", val))
	}
	return buf.String()

}
