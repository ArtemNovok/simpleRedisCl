package storage

import (
	"sync"
)

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
func (l *List) DeleteL(key []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.lists, string(key))
	return nil
}

func (l *List) DelElmL(key []byte, value []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	mapList, ok := l.lists[string(key)]
	if !ok {
		return ErrKeyDoNotExists
	}
	var newList [][]byte
	for ind, val := range mapList {
		if string(val) == string(value) {
			newList = append(mapList[:ind], mapList[ind+1:]...)
			break
		}
	}
	if len(newList) == 0 {
		delete(l.lists, string(key))
		return nil
	}
	l.lists[string(key)] = newList
	return nil
}

func (l *List) DelAll(key []byte, value []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	mapList, ok := l.lists[string(key)]
	if !ok {
		return ErrKeyDoNotExists
	}
	var newList [][]byte
	for _, val := range mapList {
		if string(val) != string(value) {
			newList = append(newList, val)
		}
	}
	if len(newList) == 0 {
		delete(l.lists, string(key))
		return nil
	}
	l.lists[string(key)] = newList
	return nil
}
