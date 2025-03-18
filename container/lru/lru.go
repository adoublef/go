// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package lru

import "container/list"

// LRU is an LRU cache. It is not safe for concurrent access.
type LRU[K comparable, V any] struct {
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	MaxEntries int

	// OnEvicted optionally specifies a callback function to be
	// executed when an entry is purged from the cache.
	OnEvicted func(key K, value V)

	ll *list.List
	m  map[K]*list.Element
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

// New creates a new [LRU].
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func New[K comparable, V any](maxEntries int) *LRU[K, V] {
	return &LRU[K, V]{
		MaxEntries: maxEntries,
		ll:         list.New(),
		m:          make(map[K]*list.Element),
	}
}

// Add adds a value to the cache.
func (c *LRU[K, V]) Add(key K, value V) {
	if c.m == nil {
		c.m = make(map[K]*list.Element)
		c.ll = list.New()
	}
	if ee, ok := c.m[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry[K, V]).value = value
		return
	}
	ele := c.ll.PushFront(&entry[K, V]{key, value})
	c.m[key] = ele
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
}

// Get looks up a key's value from the cache.
func (c *LRU[K, V]) Get(key K) (V, bool) {
	if c.m == nil {
		return *new(V), false
	}
	if ele, hit := c.m[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry[K, V]).value, true
	}
	return *new(V), false
}

// Remove removes the provided key from the cache.
func (c *LRU[K, V]) Remove(key K) {
	if c.m == nil {
		return
	}
	if ele, hit := c.m[key]; hit {
		c.removeElement(ele)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *LRU[K, V]) RemoveOldest() {
	if c.m == nil {
		return
	}
	if ele := c.ll.Back(); ele != nil {
		c.removeElement(ele)
	}
}

func (c *LRU[K, V]) removeElement(e *list.Element) {
	kv := c.ll.Remove(e).(*entry[K, V])
	delete(c.m, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len returns the number of items in the cache.
func (c *LRU[K, V]) Len() int {
	if c.m == nil {
		return 0
	}
	return c.ll.Len()
}

// Clear purges all stored items from the cache.
func (c *LRU[K, V]) Clear() {
	if c.OnEvicted != nil {
		for _, e := range c.m {
			kv := e.Value.(*entry[K, V])
			c.OnEvicted(kv.key, kv.value)
		}
	}
	c.ll = nil
	c.m = nil
}
