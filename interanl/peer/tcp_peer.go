package Mypeer

import "net"

type TCPPeer struct {
	conn net.Conn
}

func NewTCPPeer(conn net.Conn) *TCPPeer {
	return &TCPPeer{
		conn: conn,
	}
}

func (t *TCPPeer) Addr() string {
	return t.conn.RemoteAddr().String()
}

func (t *TCPPeer) ReadLoop() error {
	return nil
}
