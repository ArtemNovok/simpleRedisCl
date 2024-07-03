package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/ArtemNovok/simpleRedisCl/interanl/command"
	Mypeer "github.com/ArtemNovok/simpleRedisCl/interanl/peer"
	"github.com/ArtemNovok/simpleRedisCl/interanl/storage"
)

var (
	defaultAddress = ":6666"
	ErrUknownPeer  = errors.New("unknown peer")
)

type Config struct {
	ListenAddr string
	Log        *slog.Logger
}

type Server struct {
	Config
	peers     map[string]*Mypeer.TCPPeer
	addPeerCh chan *Mypeer.TCPPeer
	quitCh    chan struct{}
	msgCh     chan Mypeer.Message
	listener  net.Listener
	kv        *storage.KeyValue
}

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultAddress
	}
	return &Server{
		Config:    cfg,
		peers:     make(map[string]*Mypeer.TCPPeer),
		addPeerCh: make(chan *Mypeer.TCPPeer),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan Mypeer.Message),
		kv:        storage.NreKeyValue(),
	}
}

func (s *Server) ShowData() {
	s.Log.Debug("info", slog.Any("data", s.kv.Data))
}
func (s *Server) Stop() {
	close(s.quitCh)
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
		case rawMsg := <-s.msgCh:
			log.Info("got new raw message", slog.Int("bytes", len(rawMsg.Payload)))
			if err := s.handleRawMessage(rawMsg.Payload); err != nil {
				log.Error("got error while handling raw message", slog.String("error", err.Error()))
				// blup
				if peer, ok := s.peers[rawMsg.From]; ok {
					if err := binary.Write(peer.Conn, binary.BigEndian, false); err != nil {
						log.Error("failed to write response", slog.String("address", rawMsg.From))
					}
				}
				log.Error("failed to find peer with given address", slog.String("address", rawMsg.From))
			}
			if peer, ok := s.peers[rawMsg.From]; ok {
				if err := binary.Write(peer.Conn, binary.BigEndian, true); err != nil {
					log.Error("failed to write response", slog.String("address", rawMsg.From))
				}
			}
		case peer := <-s.addPeerCh:
			log.Info("added new peer", slog.String("peer address", peer.Addr()))
			s.peers[peer.Addr()] = peer
		case <-s.quitCh:
			log.Info("server stopped due to Stop func call")
			return
		}
	}
}
func (s *Server) handleRawMessage(msg []byte) error {
	const op = "server.handleRawMessage"
	log := s.Log.With("op", op)
	log.Info("start parsing the commnad")
	cmd, err := command.ParseCommand(string(msg))
	if err != nil {
		log.Error("got error while parsing command", slog.String("error", err.Error()))
		return fmt.Errorf("%s:%w", op, err)
	}
	switch v := cmd.(type) {
	case command.SetCommand:
		return s.kv.Set(v.Key, v.Val)
	}
	return nil
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
	peer := Mypeer.NewTCPPeer(conn, s.msgCh)
	s.addPeerCh <- peer
	if err := peer.ReadLoop(); err != nil {
		if errors.Is(err, io.EOF) {
			slog.Info("done handling peer", slog.String("address", conn.RemoteAddr().String()))
			return nil
		}
		slog.Error("got error while handling peer", slog.String("error", err.Error()),
			slog.String("address", conn.RemoteAddr().String()))
		return err
	}
	slog.Info("done handling peer", slog.String("address", conn.RemoteAddr().String()))
	return nil
}
