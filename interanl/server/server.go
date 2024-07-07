package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"

	"github.com/ArtemNovok/simpleRedisCl/interanl/command"
	Mypeer "github.com/ArtemNovok/simpleRedisCl/interanl/peer"
	"github.com/ArtemNovok/simpleRedisCl/interanl/storage"
)

const (
	defaultPassword = "secret"
)

var (
	DefaultAddress     = ":6666"
	ErrUknownPeer      = errors.New("unknown peer")
	ErrInvalidPassword = errors.New("invalid password")
)

type Config struct {
	ListenAddr string
	password   string
	Log        *slog.Logger
}

// Server represents goRedisClone server
type Server struct {
	Config
	mu        sync.RWMutex
	peers     map[string]*Mypeer.TCPPeer
	addPeerCh chan *Mypeer.TCPPeer
	dropPeer  chan string
	quitCh    chan struct{}
	msgCh     chan Mypeer.Message
	listener  net.Listener
	Storage   *storage.Storage
}

// NewServer returns server instance with given server Config
func NewServer(cfg Config) *Server {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = DefaultAddress
	}
	if len(cfg.password) == 0 {
		cfg.password = defaultPassword
	}
	return &Server{
		Config:    cfg,
		peers:     make(map[string]*Mypeer.TCPPeer),
		addPeerCh: make(chan *Mypeer.TCPPeer),
		dropPeer:  make(chan string),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan Mypeer.Message),
		Storage:   storage.NewStorage(),
	}
}

// ShowData shows data if log level is Debug
func (s *Server) ShowData() {
	for i := 0; i < 40; i++ {
		for key, val := range s.Storage.DBS[i].KV.Data {
			s.Log.Debug("info", slog.String("key", key), slog.String("value", string(val)))
		}
	}
}
func (s *Server) Stop() {
	close(s.quitCh)
}

// Start starts server
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
		case from := <-s.dropPeer:
			delete(s.peers, from)
		case <-s.quitCh:
			log.Info("server stopped due to Stop func call")
			return
		}
	}
}

func (s *Server) LPush(from string, key, val []byte, index int) error {
	const op = "server.LPush"
	log := s.Log.With(slog.String("op", op), slog.String("peer address", from))
	s.mu.RLock()
	peer, ok := s.peers[from]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%s:%w", op, ErrUknownPeer)
	}
	err := s.Storage.LPush(key, val, index)
	if err != nil {
		log.Error("failed to append value to a list", slog.String("key", string(key)))
		binary.Write(peer.Conn, binary.BigEndian, false)
		return fmt.Errorf("%s:%w", op, err)
	}
	binary.Write(peer.Conn, binary.BigEndian, true)
	log.Info("value is appended to a list")
	return nil
}

func (s *Server) Has(from string, key []byte, index int) error {
	const op = "server.Has"
	log := s.Log.With(slog.String("op", op), slog.String("peer address", from))
	s.mu.RLock()
	peer, ok := s.peers[from]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%s:%w", op, ErrUknownPeer)
	}
	binary.Write(peer.Conn, binary.BigEndian, s.Storage.Has(key, index))
	log.Info("Checked whether db has a key", slog.String("key", string(key)))
	return nil
}

func (s *Server) GetL(from string, key []byte, index int) error {
	const op = "server.GetL"
	log := s.Log.With(slog.String("op", op), slog.String("peer address", from))
	s.mu.RLock()
	peer, ok := s.peers[from]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%s:%w", op, ErrUknownPeer)
	}
	slice, err := s.Storage.GetL(key, index)
	if err != nil {
		binary.Write(peer.Conn, binary.BigEndian, false)
		return fmt.Errorf("%s:%w", op, storage.ErrKeyDoNotExists)
	}
	binary.Write(peer.Conn, binary.BigEndian, true)
	slLen := int64(len(slice))
	binary.Write(peer.Conn, binary.BigEndian, slLen)
	for _, sl := range slice {
		n := int64(len(sl))
		binary.Write(peer.Conn, binary.BigEndian, n)
		r := bytes.NewReader(sl)
		_, err := io.Copy(peer.Conn, r)
		if err != nil {
			return fmt.Errorf("%s:%w", op, err)
		}
	}
	log.Info("list is sended")
	return nil
}

// Set sets the key value and write response to the client with info about operation result
func (s *Server) Set(from string, key, val []byte, index int) error {
	const op = "server.Set"
	log := s.Log.With(slog.String("op", op), slog.String("peer address", from))
	s.mu.RLock()
	// we need peer to write response to client
	peer, ok := s.peers[from]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%s:%w", op, ErrUknownPeer)
	}
	err := s.Storage.Set(key, val, index)
	if err != nil {
		log.Error("failed to set a key", slog.String("key", string(key)))
		binary.Write(peer.Conn, binary.BigEndian, false)
		return fmt.Errorf("%s:%w", op, err)
	}
	binary.Write(peer.Conn, binary.BigEndian, true)
	log.Info("key is sett")
	return nil
}

// Get gets value of the key and response to the client
func (s *Server) Get(from string, key []byte, index int) error {
	const op = "server.Get"
	log := s.Log.With(slog.String("op", op), slog.String("peer address", from))
	s.mu.RLock()
	peer, ok := s.peers[from]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%s:%w", op, ErrUknownPeer)
	}
	val, ok := s.Storage.Get(key, index)
	log.Info("got value for a peer", slog.String("value", string(val)))
	if !ok {
		binary.Write(peer.Conn, binary.BigEndian, false)
		return fmt.Errorf("%s:%w", op, storage.ErrKeyDoNotExists)
	}
	n := int64(len(val))
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

// Add increments key value by 1 and writes response about success of the operation
func (s *Server) Add(from string, key []byte, index int) error {
	const op = "server.Add"
	log := s.Log.With(slog.String("op", op), slog.String("peer address", from))
	s.mu.RLock()
	peer, ok := s.peers[from]
	s.mu.RUnlock()
	if !ok {
		return ErrUknownPeer
	}
	if err := s.Storage.Add(key, index); err != nil {
		binary.Write(peer.Conn, binary.BigEndian, false)
		return fmt.Errorf("%s:%w", op, err)
	}
	binary.Write(peer.Conn, binary.BigEndian, true)
	log.Info("key value is increment by one")
	return nil
}

// AddN increments key value by given value and writes response about success of the operation
func (s *Server) AddN(from string, key []byte, value []byte, index int) error {
	const op = "server.Add"
	log := s.Log.With(slog.String("op", op), slog.String("peer address", from))
	s.mu.RLock()
	peer, ok := s.peers[from]
	s.mu.RUnlock()
	if !ok {
		return ErrUknownPeer
	}
	if err := s.Storage.AddN(key, value, index); err != nil {
		binary.Write(peer.Conn, binary.BigEndian, false)
		return fmt.Errorf("%s:%w", op, err)
	}
	binary.Write(peer.Conn, binary.BigEndian, true)
	log.Info("key value is increment by", slog.String("value", string(value)))
	return nil
}
func (s *Server) DelAll(from string, key []byte, value []byte, index int) error {
	const op = "server.DelAll"
	log := s.Log.With(slog.String("op", op), slog.String("peer address", from))
	s.mu.RLock()
	peer, ok := s.peers[from]
	s.mu.RUnlock()
	if !ok {
		return ErrUknownPeer
	}
	err := s.Storage.DelAll(key, value, index)
	if err != nil {
		binary.Write(peer.Conn, binary.BigEndian, false)
		return fmt.Errorf("%s:%w", op, err)
	}
	binary.Write(peer.Conn, binary.BigEndian, true)
	log.Info("all appearances of list value are deleted")
	return nil
}

// Delete deletes key
func (s *Server) Delete(from string, key []byte, index int) error {
	const op = "server.Delete"
	log := s.Log.With(slog.String("op", op), slog.String("peer address", from))
	s.mu.RLock()
	peer, ok := s.peers[from]
	s.mu.RUnlock()
	if !ok {
		return ErrUknownPeer
	}
	err := s.Storage.Delete(key, index)
	if err != nil {
		binary.Write(peer.Conn, binary.BigEndian, false)
		return fmt.Errorf("%s:%w", op, err)
	}
	binary.Write(peer.Conn, binary.BigEndian, true)
	log.Info("key is deleted")
	return nil
}
func (s *Server) DeleteL(from string, key []byte, index int) error {
	const op = "server.DeleteL"
	log := s.Log.With(slog.String("op", op), slog.String("peer address", from))
	s.mu.RLock()
	peer, ok := s.peers[from]
	s.mu.RUnlock()
	if !ok {
		return ErrUknownPeer
	}
	err := s.Storage.DeleteL(key, index)
	if err != nil {
		binary.Write(peer.Conn, binary.BigEndian, false)
		return fmt.Errorf("%s:%w", op, err)
	}
	binary.Write(peer.Conn, binary.BigEndian, true)
	log.Info("list is deleted")
	return nil
}
func (s *Server) DelElemL(from string, key []byte, value []byte, index int) error {
	const op = "storage.DelElemL"
	log := s.Log.With(slog.String("op", op), slog.String("peer address", from))
	s.mu.RLock()
	peer, ok := s.peers[from]
	s.mu.RUnlock()
	if !ok {
		return ErrUknownPeer
	}
	err := s.Storage.DelElemL(key, value, index)
	if err != nil {
		binary.Write(peer.Conn, binary.BigEndian, false)
		return fmt.Errorf("%s:%w", op, err)
	}
	binary.Write(peer.Conn, binary.BigEndian, true)
	log.Info("list value is deleted")
	return nil
}

// handleRawMessage handles ram message and execute logic for given type of message
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
	case command.DelAllCommnad:
		return s.DelAll(from, v.Key, v.Val, v.Index)
	case command.DelElemLCommnad:
		return s.DelElemL(from, v.Key, v.Val, v.Index)
	case command.DeleteLCommnad:
		return s.DeleteL(from, v.Key, v.Index)
	case command.LPushCommand:
		return s.LPush(from, v.Key, v.Val, v.Index)
	case command.GetLCommand:
		return s.GetL(from, v.Key, v.Index)
	case command.HasCommand:
		return s.Has(from, v.Key, v.Index)
	case command.SetCommand:
		return s.Set(from, v.Key, v.Val, v.Index)
	case command.GetCommand:
		return s.Get(from, v.Key, v.Index)
	case command.HelloCommand:
		log.Info("got hello command")
	case command.AddCommand:
		return s.Add(from, v.Key, v.Index)
	case command.AdddNCommand:
		return s.AddN(from, v.Key, v.Val, v.Index)
	case command.DeleteCommnad:
		return s.Delete(from, v.Key, v.Index)
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
	log := s.Log.With(slog.String("op", op), slog.String("connection address", conn.RemoteAddr().String()))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Error("failed to read password from peer")
		return fmt.Errorf("%s:%w", op, err)
	}
	strPass := string(buf[:n])
	if s.password != strPass {
		log.Error("peer with wrong password", slog.String("password", strPass))
		binary.Write(conn, binary.BigEndian, false)
		return fmt.Errorf("%s:%w", op, ErrInvalidPassword)
	}
	binary.Write(conn, binary.BigEndian, true)
	log.Info("starting handling connection", slog.String("address", conn.RemoteAddr().String()))
	peer := Mypeer.NewTCPPeer(conn, s.msgCh, s.dropPeer)
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
