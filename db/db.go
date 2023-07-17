package db

import (
	"fmt"
	"sync"
	"time"
)

type DB interface {
	Set([]byte, []byte, time.Duration) error
	Get([]byte) ([]byte, error)
	Store() map[string][]byte
	Clear()
}

type Cache struct {
	lock  sync.RWMutex
	store map[string][]byte
}

func New() *Cache {
	return &Cache{
		store: make(map[string][]byte),
	}
}

func (c *Cache) Get(key []byte) ([]byte, error) {
	c.lock.RLocker().Lock()
	defer c.lock.RUnlock()
	k := string(key)
	val, ok := c.store[k]
	if !ok {
		return nil, fmt.Errorf("key [%s] does not exist", k)
	}

	return val, nil

}

func (c *Cache) Set(key []byte, val []byte, duration time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	k := string(key)
	c.store[k] = val

	return nil
}

func (c *Cache) Store() map[string][]byte {
	return c.store
}

func (c *Cache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Clear the existing store
	c.store = make(map[string][]byte)
}
