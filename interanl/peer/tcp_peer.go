package Mypeer

import (
	"fmt"
	"net"
)

type TCPPeer struct {
	Conn  net.Conn
	msgCh chan Message
}
type Message struct {
	From    string
	Payload []byte
}

func NewTCPPeer(conn net.Conn, msgch chan Message) *TCPPeer {
	return &TCPPeer{
		Conn:  conn,
		msgCh: msgch,
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
