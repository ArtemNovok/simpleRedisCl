package storage

import (
	"errors"
	"strconv"
	"sync"
)

var (
	ErrKeyDoNotExists       = errors.New("key doesn't exist")
	ErrUnableToConvertToInt = errors.New("unable to convert value to integer")
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

func (kv *KeyValue) Add(key []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	val, ok := kv.Data[string(key)]
	if !ok {
		return ErrKeyDoNotExists
	}
	intVal, err := strconv.Atoi(string(val))
	if err != nil {
		return ErrUnableToConvertToInt
	}
	intVal++

	kv.Data[string(key)] = []byte(strconv.Itoa(intVal))
	return nil
}
func (kv *KeyValue) AddN(key []byte, value []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	val, ok := kv.Data[string(key)]
	if !ok {
		return ErrKeyDoNotExists
	}
	intVal, err := strconv.Atoi(string(val))
	if err != nil {
		return ErrUnableToConvertToInt
	}
	intV, err := strconv.Atoi(string(value))
	if err != nil {
		return ErrUnableToConvertToInt
	}
	intVal += intV

	kv.Data[string(key)] = []byte(strconv.Itoa(intVal))
	return nil
}
func (kv *KeyValue) Delete(key []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	delete(kv.Data, string(key))
	return nil
}
