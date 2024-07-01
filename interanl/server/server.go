package server

import (
	"fmt"
	"log/slog"
	"net"

	Mypeer "github.com/ArtemNovok/simpleRedisCl/interanl/peer"
)

const (
	defaultAddress = ":6666"
)

type Config struct {
	ListenAddr string
	Log        *slog.Logger
}

type Server struct {
	Config
	peers     map[*Mypeer.TCPPeer]bool
	addPeerCh chan *Mypeer.TCPPeer
	listener  net.Listener
}

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultAddress
	}
	return &Server{
		Config:    cfg,
		peers:     make(map[*Mypeer.TCPPeer]bool),
		addPeerCh: make(chan *Mypeer.TCPPeer),
	}
}

func (s *Server) Start() error {
	const op = "server.Start"
	log := s.Log.With("op", op)
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	log.Info("starting listening", slog.String("address", s.ListenAddr))
	s.listener = ln
	go s.loop()
	return s.listenLoop()
}

func (s *Server) loop() {
	const op = "server.loop"
	log := s.Log.With("op", op)
	for {
		select {
		case peer := <-s.addPeerCh:
			log.Info("added new peer", slog.String("peer address", peer.Addr()))
			s.peers[peer] = true
		}
	}
}

func (s *Server) listenLoop() error {
	const op = "server.listenLoop"
	log := s.Log.With("op", op)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Error("got error while accepting connections", slog.String("error", err.Error()))
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) error {
	const op = "server.handleConn"
	log := s.Log.With("op", op)
	log.Info("starting handling connection", slog.String("address", conn.RemoteAddr().String()))
	peer := Mypeer.NewTCPPeer(conn)
	s.addPeerCh <- peer
	go peer.ReadLoop()
	return nil
}
