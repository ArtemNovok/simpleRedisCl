package client

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"

	"github.com/tidwall/resp"
)

const (
	defaultPassword = "secret"
)

var (
	CommandDelete   = "DEL"
	CommandSet      = "SET"
	CommandGet      = "GET"
	CommandHello    = "HELLO"
	CommandAdd      = "ADD"
	CommandAddN     = "ADDN"
	CommandLPush    = "LPUSH"
	CommandGetL     = "GETL"
	CommandHas      = "HAS"
	CommandDeleteL  = "DELL"
	CommandDelElemL = "DELELEML"
	CommandDelAll   = "DELALL"
	// ErrOperationFailed returned when operation failed not due to context cancel
	ErrOperationFailed = errors.New("operation failed")
	// ErrTimeIsOut returned when operation failed due to context cancel
	ErrTimeIsOut = errors.New("time is out")
	// ErrInvalidIndex returned when operation index value  beyond  0 <= ind <= 39
	ErrInvalidIndex = errors.New("invalid index value")
	// ErrInvalidPassword returned when wrong password is used to connect to a server
	ErrInvalidPassword = errors.New("invalid password")
)

// Client used for communication between app and server, it supports concurrent operations
type Client struct {
	addr     string
	connLock sync.Mutex
	conn     net.Conn
	password string
}
type GetResult struct {
	value string
	err   error
}

// New create connection  to the server and returns client with that connection and  error if occurs
func New(ctx context.Context, addr string, password string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	if len(password) == 0 {
		password = defaultPassword
	}
	_, err = conn.Write([]byte(password))
	if err != nil {
		return nil, err
	}
	ch := make(chan bool)
	go func() {
		var res bool
		err = binary.Read(conn, binary.BigEndian, &res)
		if err != nil {
			ch <- false

		} else {
			ch <- res
		}
	}()
	select {
	case <-ctx.Done():
		return nil, ErrTimeIsOut
	case res := <-ch:
		if !res {
			return nil, ErrInvalidPassword
		}
		return &Client{
			addr:     addr,
			conn:     conn,
			password: password,
		}, nil
	}
}
func (c *Client) writeRequest(cmd string, ind int, args ...string) error {
	index := strconv.Itoa(ind)
	respReq := []resp.Value{resp.StringValue(cmd)}
	for _, val := range args {
		respReq = append(respReq, resp.StringValue(val))
	}
	respReq = append(respReq, resp.StringValue(index))
	buf := &bytes.Buffer{}
	wr := resp.NewWriter(buf)
	err := wr.WriteArray(respReq)
	if err != nil {
		return err
	}
	_, err = io.Copy(c.conn, buf)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) readResponse(ch chan error) {
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
}
func (c *Client) waitForResponse(ch chan error, ctx context.Context) error {
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
func (c *Client) DelAll(ctx context.Context, key string, value string, ind int) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if ind > 39 && ind < 0 {
		return ErrInvalidIndex
	}
	if err := c.writeRequest(CommandDelAll, ind, key, value); err != nil {
		return err
	}
	ch := make(chan error)
	go c.readResponse(ch)
	return c.waitForResponse(ch, ctx)
}
func (c *Client) DelElemL(ctx context.Context, key string, value string, ind int) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if ind > 39 && ind < 0 {
		return ErrInvalidIndex
	}
	if err := c.writeRequest(CommandDelElemL, ind, key, value); err != nil {
		return err
	}
	ch := make(chan error)
	go c.readResponse(ch)
	return c.waitForResponse(ch, ctx)
}

func (c *Client) DeleteL(ctx context.Context, key string, ind int) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if ind > 39 && ind < 0 {
		return ErrInvalidIndex
	}
	if err := c.writeRequest(CommandDeleteL, ind, key); err != nil {
		return err
	}
	ch := make(chan error)
	go c.readResponse(ch)
	return c.waitForResponse(ch, ctx)
}

func (c *Client) Delete(ctx context.Context, key string, ind int) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if ind > 39 && ind < 0 {
		return ErrInvalidIndex
	}
	if err := c.writeRequest(CommandDelete, ind, key); err != nil {
		return err
	}
	ch := make(chan error)
	go c.readResponse(ch)
	return c.waitForResponse(ch, ctx)
}
func (c *Client) GetL(ctx context.Context, key string, ind int) ([]string, error) {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if ind > 39 && ind < 0 {
		return nil, ErrInvalidIndex
	}
	if err := c.writeRequest(CommandGetL, ind, key); err != nil {
		return nil, err
	}
	ch := make(chan []string)
	go func() {
		var res bool
		binary.Read(c.conn, binary.BigEndian, &res)
		if !res {
			ch <- nil
			return
		}
		var lenSL int64
		binary.Read(c.conn, binary.BigEndian, &lenSL)
		slice := make([]string, 0, lenSL)
		for i := 0; i < int(lenSL); i++ {
			var size int64
			err := binary.Read(c.conn, binary.BigEndian, &size)
			if err != nil {
				ch <- nil
				return
			}
			buf := make([]byte, size)
			if _, err := c.conn.Read(buf); err != nil {
				ch <- nil
				return
			}
			slice = append(slice, string(buf))
		}
		ch <- slice
	}()

	select {
	case <-ctx.Done():
		return nil, ErrTimeIsOut
	case res := <-ch:
		if res == nil {
			return nil, ErrOperationFailed
		}
		return res, nil
	}
}
func (c *Client) Has(ctx context.Context, key string, ind int) (bool, error) {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if ind > 39 && ind < 0 {
		return false, ErrInvalidIndex
	}
	if err := c.writeRequest(CommandHas, ind, key); err != nil {
		return false, err
	}
	ch := make(chan bool)
	go func() {
		var res bool
		err := binary.Read(c.conn, binary.BigEndian, &res)
		if err != nil {
			ch <- false
			return
		}
		ch <- res
	}()
	select {
	case <-ctx.Done():
		return false, ErrTimeIsOut
	case res := <-ch:
		return res, nil
	}

}
func (c *Client) LPush(ctx context.Context, key string, value string, ind int) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if ind > 39 && ind < 0 {
		return ErrInvalidIndex
	}
	if err := c.writeRequest(CommandLPush, ind, key, value); err != nil {
		return err
	}
	ch := make(chan error)
	go c.readResponse(ch)
	return c.waitForResponse(ch, ctx)
}

// Set sets key with given value it returns error if ctx is done or operation failed
func (c *Client) Set(ctx context.Context, key string, value string, ind int) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if ind > 39 && ind < 0 {
		return ErrInvalidIndex
	}
	if err := c.writeRequest(CommandSet, ind, key, value); err != nil {
		return err
	}
	ch := make(chan error)
	go c.readResponse(ch)
	return c.waitForResponse(ch, ctx)
}

// Get reruns key value and  error if ctx is done or operation failed
func (c *Client) Get(ctx context.Context, key string, ind int) (string, error) {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if ind > 39 && ind < 0 {
		return "", ErrInvalidIndex
	}
	if err := c.writeRequest(CommandGet, ind, key); err != nil {
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
func (c *Client) Add(ctx context.Context, key string, ind int) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if ind > 39 && ind < 0 {
		return ErrInvalidIndex
	}
	if err := c.writeRequest(CommandAdd, ind, key); err != nil {
		return err
	}
	ch := make(chan error)
	go c.readResponse(ch)
	return c.waitForResponse(ch, ctx)
}

// AddN increment key value by given value and  returns error if ctx is done or operation failed
func (c *Client) AddN(ctx context.Context, key string, value string, ind int) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if ind > 39 && ind < 0 {
		return ErrInvalidIndex
	}
	if err := c.writeRequest(CommandAddN, ind, key, value); err != nil {
		return err
	}
	ch := make(chan error)
	go c.readResponse(ch)
	return c.waitForResponse(ch, ctx)
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
