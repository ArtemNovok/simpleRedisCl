package Mypeer

import (
	"fmt"
	"net"
)

type TCPPeer struct {
	conn  net.Conn
	msgCh chan []byte
}

func NewTCPPeer(conn net.Conn, msgch chan []byte) *TCPPeer {
	return &TCPPeer{
		conn:  conn,
		msgCh: msgch,
	}
}

func (t *TCPPeer) Addr() string {
	return t.conn.RemoteAddr().String()
}

func (t *TCPPeer) ReadLoop() error {
	const op = "peer.ReadLoop"
	buf := make([]byte, 1024)
	for {
		n, err := t.conn.Read(buf)
		if err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}
		msgBuf := make([]byte, n)
		copy(msgBuf, buf[:n])
		t.msgCh <- msgBuf
	}
}
