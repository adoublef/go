// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package hashqueue provides a serialisation function call execution
// mechanism.
package hashqueue

import (
	"sync"

	"go.adoublef.dev/runtime/debug"
)

type call struct {
	funcs chan func()
	count int64
}

type Group[K comparable] struct {
	mu sync.Mutex
	m  map[K]*call
}

func (g *Group[K]) Do(key K, f func() error) error {
	c := g.loadCall(key)

	res := make(chan error, 1)
	c.funcs <- func() {
		res <- f()
	}
	return <-res
}

func (g *Group[K]) DoChan(key K, f func() error) <-chan error {
	c := g.loadCall(key)

	res := make(chan error, 1)
	c.funcs <- func() {
		res <- f()
	}
	return res
}

func (g *Group[K]) TryDo(key K, f func() error) (error, bool) {
	c := g.loadCall(key)

	res := make(chan error, 1)
	select {
	case c.funcs <- func() {
		res <- f()
	}:
		return <-res, true
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
		return nil, false
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
	defer debug.Printf("closing call for key %v", key)
	normalReturn := false

	for f := range c.funcs {
		f()

		func() {
			g.mu.Lock()
			defer g.mu.Unlock()

			if c.count--; c.count == 0 {
				delete(g.m, key)
				normalReturn = true
			}
		}()

		if normalReturn {
			return
		}
	}
}
