// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package batchque

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type Group[K comparable, V any] struct {
	BatchTimeout time.Duration
	wg           sync.WaitGroup
	init         sync.Once
	requests     chan *Request[K, V]
	quit         chan struct{}
	closed       atomic.Bool
}

func (g *Group[K, V]) Do(ctx context.Context, key K, f func(context.Context, []*Request[K, V])) (V, error) {
	if g.closed.Load() {
		return *new(V), errors.New("group is closed") // make meaningful
	}
	g.init.Do(runLoop(g, f))

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	res := make(chan V, 1)
	r := &Request[K, V]{
		Val:        key,
		C:          res,
		ctx:        ctx,
		CancelFunc: cancel,
	}

	select {
	case g.requests <- r: // was able to put it on the batch queue
	case <-ctx.Done():
		return *new(V), context.Cause(ctx)
	}
	select {
	case res := <-res:
		return res, nil
	case <-ctx.Done():
		return *new(V), context.Cause(ctx)
	}
}

func runLoop[K comparable, V any](g *Group[K, V], f func(context.Context, []*Request[K, V])) func() {
	return func() {
		g.requests = make(chan *Request[K, V], 16)
		g.quit = make(chan struct{})

		g.wg.Add(1)
		go func() {
			defer g.wg.Done()

			rr := make([]*Request[K, V], 0, 1<<10) // =1kb
			for {
				select {
				case r := <-g.requests:
					rr = append(rr, r)
				EMPTY:
					for {
						select {
						case r := <-g.requests:
							rr = append(rr, r)
						default:
							break EMPTY
						}
					}
					// TOOD: add a timeout
					ctx, cancel := context.WithCancel(context.Background())
					f(ctx, rr)
					cancel()
					rr = rr[:0]
				case <-g.quit:
					return
				}
			}
		}()
	}
}

func (g *Group[K, V]) Close() error {
	if g.closed.CompareAndSwap(false, true) {
		if g.quit != nil {
			close(g.quit)
		}
		g.wg.Wait()
	}
	return nil
}

// Request is just a channel absctraction
type Request[V, R any] struct {
	Val        V
	C          chan<- R
	ctx        context.Context
	CancelFunc context.CancelCauseFunc
}

func (r Request[V, R]) Context() context.Context {
	if r.ctx == nil {
		return context.Background()
	}
	return r.ctx
}
