// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cache provides a generic in-memory LRU (Least Recently Used) cache implementation.
// The cache stores any Go value by serializing it to bytes using MessagePack encoding,
// making it suitable for a wide variety of data types.
package cache

import (
	"fmt"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
	"go.adoublef.dev/container/lru"
)

type Key interface {
	comparable
	~string | ~[]byte | ~[12]byte | ~[16]byte
}

// LRU is a generic LRU cache that can store any value type.
type LRU[K Key, V any] struct {
	mu         sync.RWMutex
	nbytes     int64               // of all keys and values
	lru        *lru.LRU[K, []byte] // any value should be allowed
	nhit, nget int64
	nevict     int64 // number of evictions
}

// Add inserts a key-value pair into the cache. If the key already exists, its value
// is updated. Returns an error if the value cannot be serialized.
func (c *LRU[K, V]) Add(key K, value V) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		// probs worth setting the [sync.Once] here?
		c.lru = &lru.LRU[K, []byte]{
			OnEvicted: func(key K, p []byte) {
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

// Get retrieves a value from the cache by its key. Returns the value and a boolean
// indicating whether the key was found.
func (c *LRU[K, V]) Get(key K) (value V, ok bool) {
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

// Remove deletes an item from the cache by its key.
func (c *LRU[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru != nil {
		c.lru.Remove(key)
	}
}

// RemoveOldest removes the least recently used item from the cache.
func (c *LRU[K, V]) RemoveOldest() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru != nil {
		c.lru.RemoveOldest()
	}
}

// Bytes returns the approximate memory usage of the cache in bytes, including both
// keys and serialized values.
func (c *LRU[K, V]) Bytes() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// if this is generic can we still get the size of the cache?
	return c.nbytes
}

// Items returns the number of items currently stored in the cache.
func (c *LRU[K, V]) Items() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.itemsLocked()
}

func (c *LRU[K, V]) itemsLocked() int64 {
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
