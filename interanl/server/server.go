package server

import (
	"bytes"
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
	DefaultAddress = ":6666"
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
		cfg.ListenAddr = DefaultAddress
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
	for key, val := range s.kv.Data {
		s.Log.Debug("info", slog.String("key", key), slog.String("value", string(val)))
	}
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
			if err := s.handleRawMessage(rawMsg.From, rawMsg.Payload); err != nil {
				log.Error("got error while handling raw message", slog.String("error", err.Error()))
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
func (s *Server) Set(from string, key, val []byte) error {
	const op = "server.Set"
	log := s.Log.With(slog.String("op", op), slog.String("peer address", from))
	// we need peer to write response to client
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("%s:%w", op, ErrUknownPeer)
	}
	err := s.kv.Set(key, val)
	if err != nil {
		log.Error("failed to set a key", slog.String("key", string(key)))
		binary.Write(peer.Conn, binary.BigEndian, false)
		return fmt.Errorf("%s:%w", op, err)
	}
	binary.Write(peer.Conn, binary.BigEndian, true)
	log.Info("key is sett")
	return nil
}

func (s *Server) Get(from string, key []byte) error {
	const op = "server.Get"
	log := s.Log.With(slog.String("op", op), slog.String("peer address", from))
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("%s:%w", op, ErrUknownPeer)
	}
	val, ok := s.kv.Get(key)
	n := int64(len(val))
	if !ok {
		binary.Write(peer.Conn, binary.BigEndian, false)
		return fmt.Errorf("%s:%w", op, storage.ErrKeyDoNotExists)
	}
	binary.Write(peer.Conn, binary.BigEndian, true)
	binary.Write(peer.Conn, binary.BigEndian, n)
	r := bytes.NewReader(val)
	_, err := io.Copy(peer.Conn, r)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	log.Info("key value is find and sended to peer")
	return nil
}

func (s *Server) handleRawMessage(from string, msg []byte) error {
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
		return s.Set(from, v.Key, v.Val)
	case command.GetCommand:
		return s.Get(from, v.Key)
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
