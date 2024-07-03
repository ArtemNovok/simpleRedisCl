package storage

import (
	"errors"
	"sync"
)

var (
	ErrKeyDoNotExists = errors.New("key doesn't exist")
)

type KeyValue struct {
	mu   sync.RWMutex
	Data map[string][]byte
}

func NreKeyValue() *KeyValue {
	return &KeyValue{
		Data: make(map[string][]byte),
	}
}

func (kv *KeyValue) Set(key, val []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.Data[string(key)] = []byte(val)
	return nil
}
func (kv *KeyValue) Get(key []byte) ([]byte, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	val, err := kv.Data[string(key)]
	return val, err
}
