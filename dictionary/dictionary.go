package dictionary

import (
	"sync"
)

// Dictionary - the dictionary object with key of type uint32 & vlaue of type []byte
type Dictionary struct {
	items map[uint32][]byte
	lock  sync.RWMutex
}

// Add adds a new item to the dictionary
func (dict *Dictionary) Add(key uint32, value []byte) {
	dict.lock.Lock()
	defer dict.lock.Unlock()
	if dict.items == nil {
		dict.items = make(map[uint32][]byte)
	}
	dict.items[key] = value
}

// Remove removes a value from the dictionary, given its key
func (dict *Dictionary) Remove(key uint32) bool {
	dict.lock.Lock()
	defer dict.lock.Unlock()
	_, ok := dict.items[key]
	if ok {
		delete(dict.items, key)
	}
	return ok
}

// Exist returns true if the key exists in the dictionary
func (dict *Dictionary) Exist(key uint32) bool {
	dict.lock.RLock()
	defer dict.lock.RUnlock()
	_, ok := dict.items[key]
	return ok
}

// Get returns the value associated with the key
func (dict *Dictionary) Get(key uint32) []byte {
	dict.lock.RLock()
	defer dict.lock.RUnlock()
	return dict.items[key]
}

// Clear removes all the items from the dictionary
func (dict *Dictionary) Clear() {
	dict.lock.Lock()
	defer dict.lock.Unlock()
	dict.items = make(map[uint32][]byte)
}

// Size returns the amount of elements in the dictionary
func (dict *Dictionary) Size() int {
	dict.lock.RLock()
	defer dict.lock.RUnlock()
	return len(dict.items)
}

// GetKeys returns a slice of all the keys present
func (dict *Dictionary) GetKeys() []uint32 {
	dict.lock.RLock()
	defer dict.lock.RUnlock()
	keys := []uint32{}
	for i := range dict.items {
		keys = append(keys, i)
	}
	return keys
}

// GetValues returns a slice of all the values present
func (dict *Dictionary) GetValues() [][]byte {
	dict.lock.RLock()
	defer dict.lock.RUnlock()
	values := [][]byte{}
	for i := range dict.items {
		values = append(values, dict.items[i])
	}
	return values
}
