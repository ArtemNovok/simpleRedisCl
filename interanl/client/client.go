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
	CommandDelete = "DEL"
	CommandSet    = "SET"
	CommandGet    = "GET"
	CommnadHello  = "HELLO"
	CommandAdd    = "ADD"
	CommandAddN   = "ADDN"
	// ErrOperationFailed returned when operation failed not due to context cancel
	ErrOperationFailed = errors.New("operation failed")
	// ErrTimeIsOut returned whe operation failed due to context cancel
	ErrTimeIsOut = errors.New("time is out")
)

// Client used for communication between app and server, it supports concurrent operations
type Client struct {
	addr     string
	connLock sync.Mutex
	conn     net.Conn
}
type GetResult struct {
	value string
	err   error
}

// New create connection  to the server and returns client with that connection and  error if occurs
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
func (c *Client) Delete(ctx context.Context, key string) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	buf := &bytes.Buffer{}
	wr := resp.NewWriter(buf)
	err := wr.WriteArray([]resp.Value{resp.StringValue(CommandDelete),
		resp.StringValue(key),
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
		ch <- err
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

// Set sets key with given value it returns error if ctx is done or operation failed
func (c *Client) Set(ctx context.Context, key string, value string) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	buf := &bytes.Buffer{}
	wr := resp.NewWriter(buf)
	err := wr.WriteArray([]resp.Value{resp.StringValue(CommandSet),
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
			return
		}
		if !res {
			ch <- ErrOperationFailed
			return
		}
		ch <- nil
		return

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

// Get reruns key value and  error if ctx is done or operation failed

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	buf := &bytes.Buffer{}
	wr := resp.NewWriter(buf)
	err := wr.WriteArray([]resp.Value{resp.StringValue(CommandGet),
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

// Add increment key value by 1 and  returns error if ctx is done or operation failed
func (c *Client) Add(ctx context.Context, key string) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	buf := &bytes.Buffer{}
	wr := resp.NewWriter(buf)
	err := wr.WriteArray([]resp.Value{resp.StringValue(CommandAdd),
		resp.StringValue(key),
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
		err := binary.Read(c.conn, binary.BigEndian, &res)
		if err != nil {
			ch <- err
			return
		}
		if !res {
			ch <- ErrOperationFailed
			return
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

// AddN increment key value by given value and  returns error if ctx is done or operation failed
func (c *Client) AddN(ctx context.Context, key string, value string) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	buf := &bytes.Buffer{}
	wr := resp.NewWriter(buf)
	err := wr.WriteArray([]resp.Value{resp.StringValue(CommandAddN),
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
		err := binary.Read(c.conn, binary.BigEndian, &res)
		if err != nil {
			ch <- err
			return
		}
		if !res {
			ch <- ErrOperationFailed
			return
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
