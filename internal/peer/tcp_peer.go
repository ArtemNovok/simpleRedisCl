package Mypeer

import (
	"errors"
	"fmt"
	"io"
	"net"
)

type TCPPeer struct {
	Conn   net.Conn
	msgCh  chan Message
	dropCh chan string
}
type Message struct {
	From    string
	Payload []byte
}

func NewTCPPeer(conn net.Conn, msgch chan Message, dropCh chan string) *TCPPeer {
	return &TCPPeer{
		Conn:   conn,
		msgCh:  msgch,
		dropCh: dropCh,
	}
}

func (t *TCPPeer) Addr() string {
	return t.Conn.RemoteAddr().String()
}

func (t *TCPPeer) ReadLoop() error {
	const op = "peer.ReadLoop"
	buf := make([]byte, 1024)
	for {
		n, err := t.Conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				t.dropCh <- t.Addr()
				return fmt.Errorf("%s:%w", op, err)
			}
			return fmt.Errorf("%s:%w", op, err)
		}
		msgBuf := make([]byte, n)
		copy(msgBuf, buf[:n])
		msg := Message{
			From:    t.Addr(),
			Payload: msgBuf,
		}
		t.msgCh <- msg
	}
}
