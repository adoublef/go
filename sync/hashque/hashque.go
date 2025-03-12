// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package hashqueue provides a serialisation function call execution
// mechanism.
package hashque

import (
	"sync"

	"go.adoublef.dev/runtime/debug"
)

func ValueFunc[K comparable, T any](g *Group[K], key K, f func() T) T {
	c := make(chan T)
	g.Do(key, func() {
		defer close(c)
		c <- f()
	})
	return <-c
}

func TryValueFunc[K comparable, V any](g *Group[K], key K, f func() V) (V, bool) {
	c := make(chan V, 1)
	ok := g.TryDo(key, func() {
		defer close(c)
		c <- f()
	})
	return <-c, ok
}

type Result[V any] struct {
	Val V
	Err error
}

func ResultFunc[K comparable, T any](g *Group[K], key K, f func() (T, error)) (T, error) {
	c := make(chan Result[T])
	g.Do(key, func() {
		defer close(c)
		var r Result[T]
		r.Val, r.Err = f()
		c <- r
	})
	r := <-c
	return r.Val, r.Err
}

func ResultChan[K comparable, V any](g *Group[K], key K, f func() (V, error)) <-chan Result[V] {
	// should maybe add a way to cancel this?
	res := make(chan Result[V], 1)
	g.Do(key, func() {
		defer close(res)
		var r Result[V]
		r.Val, r.Err = f()
		res <- r
	})
	return res
}

type call struct {
	funcs chan func()
	count int64
}

type Group[K comparable] struct {
	mu sync.Mutex
	m  map[K]*call
}

func (g *Group[K]) Do(key K, f func()) {
	c := g.loadCall(key)

	c.funcs <- f
}

func (g *Group[K]) TryDo(key K, f func()) bool {
	c := g.loadCall(key)

	select {
	case c.funcs <- f:
		return true
	default:
		func() {
			g.mu.Lock()
			defer g.mu.Unlock()

			if c.count--; c.count == 0 {
				// we're the last waiter therefore
				// closing the channel is ok.
				delete(g.m, key)
				close(c.funcs)
			}
		}()
		return false
	}
}

func (g *Group[K]) loadCall(key K) *call {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[K]*call)
	}
	c, ok := g.m[key]
	if ok {
		c.count++
		g.mu.Unlock()
	} else {
		c = &call{
			funcs: make(chan func(), 16),
		}
		c.count++
		g.m[key] = c
		g.mu.Unlock()

		go g.doCall(c, key)
	}
	return c
}

func (g *Group[K]) doCall(c *call, key K) {
	defer debug.Printf("sync/hashqueue: closing call for key %v", key)
	normalReturn := false

	for f := range c.funcs {
		f()

		func() {
			g.mu.Lock()
			defer g.mu.Unlock()

			if c.count--; c.count == 0 {
				delete(g.m, key)
				// closing the channel not needed
				// but look to see if I can.
				normalReturn = true
			}
		}()

		if normalReturn {
			return
		}
	}
}
