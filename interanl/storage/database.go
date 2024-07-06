package storage

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidDatabaseIndex = errors.New("invalid data base index")
)

type DataBase struct {
	Index int
	KV    *KeyValue
	LST   *List
}
type Storage struct {
	DBS [40]*DataBase
}

func NewStorage() *Storage {
	s := Storage{}
	for i := 0; i < 40; i++ {
		db := DataBase{
			Index: i,
			KV:    NreKeyValue(),
			LST:   NewList(),
		}
		s.DBS[i] = &db
	}
	return &s
}

func (s *Storage) Set(key []byte, value []byte, index int) error {
	const op = "storage.Set"
	if index > 39 || index < 0 {
		return fmt.Errorf("%s:%w", op, ErrInvalidDatabaseIndex)
	}
	err := s.DBS[index].KV.Set(key, value)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	return nil
}

func (s *Storage) Get(key []byte, index int) ([]byte, bool) {
	if index > 39 || index < 0 {
		return nil, false
	}
	return s.DBS[index].KV.Get(key)
}

func (s *Storage) Add(key []byte, index int) error {
	const op = "storage.Add"
	if index > 39 || index < 0 {
		return fmt.Errorf("%s:%w", op, ErrInvalidDatabaseIndex)
	}
	err := s.DBS[index].KV.Add(key)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	return nil
}

func (s *Storage) AddN(key []byte, value []byte, index int) error {
	const op = "storage.AddN"
	if index > 39 || index < 0 {
		return fmt.Errorf("%s:%w", op, ErrInvalidDatabaseIndex)
	}
	err := s.DBS[index].KV.AddN(key, value)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	return nil
}

func (s *Storage) Delete(key []byte, index int) error {
	const op = "storage.Delete"
	if index > 39 || index < 0 {
		return fmt.Errorf("%s:%w", op, ErrInvalidDatabaseIndex)
	}
	return s.DBS[index].KV.Delete(key)
}

func (s *Storage) LPush(key []byte, value []byte, index int) error {
	const op = "storage.LPush"
	if index > 39 || index < 0 {
		return fmt.Errorf("%s:%w", op, ErrInvalidDatabaseIndex)
	}
	return s.DBS[index].LST.LPush(key, value)
}
func (s *Storage) Has(key []byte, index int) bool {
	if index > 39 || index < 0 {
		return false
	}
	return s.DBS[index].LST.Has(key)
}
func (s *Storage) GetL(key []byte, index int) ([][]byte, error) {
	const op = "storage.GetL"
	if index > 39 || index < 0 {
		return nil, fmt.Errorf("%s:%w", op, ErrInvalidDatabaseIndex)
	}
	return s.DBS[index].LST.GetL(key)
}

func (s *Storage) DeleteL(key []byte, index int) error {
	const op = "storage.DeleteL"
	if index > 39 || index < 0 {
		return fmt.Errorf("%s:%w", op, ErrInvalidDatabaseIndex)
	}
	return s.DBS[index].LST.DeleteL(key)
}

func (s *Storage) DelElemL(key []byte, value []byte, index int) error {
	const op = "storage.DelElemL"
	if index > 39 || index < 0 {
		return fmt.Errorf("%s:%w", op, ErrInvalidDatabaseIndex)
	}
	return s.DBS[index].LST.DelElmL(key, value)
}
