package client

import (
	"bytes"
	"context"
	"net"

	"github.com/tidwall/resp"
)

type Client struct {
	addr string
}

func New(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

func (c *Client) Set(ctx context.Context, key string, value string) error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	var buf bytes.Buffer
	wr := resp.NewWriter(&buf)
	err = wr.WriteArray([]resp.Value{resp.StringValue("SET"),
		resp.StringValue(key),
		resp.StringValue(value),
	})
	if err != nil {
		return nil
	}
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
