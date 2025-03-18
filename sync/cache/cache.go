// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"fmt"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
	"go.adoublef.dev/container/lru"
)

type Cache[V any] struct {
	mu         sync.RWMutex
	nbytes     int64                    // of all keys and values
	lru        *lru.LRU[string, []byte] // any value should be allowed
	nhit, nget int64
	nevict     int64 // number of evictions
}

func (c *Cache[V]) Add(key string, value V) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		// probs worth setting the [sync.Once] here?
		c.lru = &lru.LRU[string, []byte]{
			OnEvicted: func(key string, p []byte) {
				c.nbytes -= int64(len(key)) + int64(len(p)) // ?
				c.nevict++
			},
		}
	}
	p, err := marshal(value)
	if err != nil {
		return fmt.Errorf("cannot convert struct to bytes slice: %w", err)
	}
	c.lru.Add(key, p)
	c.nbytes += int64(len(key)) + int64(len(p)) // is this
	return nil
}

func (c *Cache[V]) Get(key string) (value V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nget++
	if c.lru == nil {
		return
	}
	// todo: batch get?
	vi, ok := c.lru.Get(key)
	if !ok {
		return
	}
	c.nhit++
	value, _ = unmarshal[V](vi)
	return value, true
}

func (c *Cache[V]) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru != nil {
		c.lru.Remove(key)
	}
}

func (c *Cache[V]) RemoveOldest() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru != nil {
		c.lru.RemoveOldest()
	}
}

func (c *Cache[V]) Bytes() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// if this is generic can we still get the size of the cache?
	return c.nbytes
}

func (c *Cache[V]) Items() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.itemsLocked()
}

func (c *Cache[V]) itemsLocked() int64 {
	if c.lru == nil {
		return 0
	}
	return int64(c.lru.Len())
}

func marshal[V any](value V) ([]byte, error) {
	p, err := msgpack.Marshal(value)
	return p, err
}

func unmarshal[V any](p []byte) (v V, err error) {
	err = msgpack.Unmarshal(p, &v)
	return v, err
}
