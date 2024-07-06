package storage

import "sync"

type List struct {
	mu    sync.Mutex
	lists map[string][][]byte
}

func NewList() *List {
	return &List{
		lists: make(map[string][][]byte),
	}
}

func (l *List) LPush(key []byte, value []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	mapList, ok := l.lists[string(key)]
	if !ok {
		temp := [][]byte{value}
		l.lists[string(key)] = temp
		return nil
	}
	mapList = append(mapList, value)
	l.lists[string(key)] = mapList
	return nil
}

func (l *List) Has(key []byte) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, ok := l.lists[string(key)]
	return ok
}

func (l *List) GetL(key []byte) ([][]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	mapList, ok := l.lists[string(key)]
	if !ok {
		return nil, ErrKeyDoNotExists
	}
	return mapList, nil
}
