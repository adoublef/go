// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package batchque

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Group[K comparable, V any] struct {
	BatchTimeout time.Duration
	wg           sync.WaitGroup
	init         sync.Once
	requests     chan *Request[K, V]
	close        sync.Once
	quit         chan struct{}
	closed       atomic.Bool
}

func (g *Group[K, V]) Do(ctx context.Context, key K, f func(context.Context, []*Request[K, V])) (V, error) {
	g.init.Do(func() {
		if !g.closed.Load() {
			g.requests = make(chan *Request[K, V], 16)
			g.quit = make(chan struct{})

			g.wg.Add(1)
			go g.runLoop(f)
		}
	})
	if g.closed.Load() {
		return *new(V), context.DeadlineExceeded
	}

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	res := make(chan V, 1)
	r := &Request[K, V]{
		Val:        key,
		C:          res,
		CancelFunc: cancel,
	}
	select {
	case <-ctx.Done():
		return *new(V), context.Cause(ctx)
	case g.requests <- r: // was able to put it on the batch queue
	}
	select {
	case <-ctx.Done():
		r.cancelled.Store(true)
		return *new(V), context.Cause(ctx)
	case res := <-res:
		return res, nil
	}
}

func (g *Group[K, V]) runLoop(f func(context.Context, []*Request[K, V])) {
	defer g.wg.Done()

	rr := make([]*Request[K, V], 1<<10) // =1kb
	var idx int
	for {
		select {
		case a := <-g.requests:
			rr[idx] = a
			idx++
		EMPTY:
			for {
				select {
				case a = <-g.requests:
					rr[idx] = a
					idx++
				default:
					break EMPTY
				}
			}
			ctx, cancel := context.WithCancel(context.Background())
			f(ctx, rr[:idx])
			cancel()
			idx = 0
		case <-g.quit:
			return
		}
	}
}

func (g *Group[K, V]) Close() error {
	g.close.Do(func() {
		if g.quit != nil {
			close(g.quit)
		}
		g.wg.Wait()
	})
	return nil
}

// Request is just a channel absctraction
type Request[V, R any] struct {
	Val        V
	CancelFunc context.CancelCauseFunc
	cancelled  atomic.Bool
	C          chan<- R
}

func (r *Request[V, R]) Closed() bool {
	return r.cancelled.Load()
}
