// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package batchque

import (
	"context"
	"sync"
	"sync/atomic"
)

type Group[K comparable, V any] struct {
	wg       sync.WaitGroup
	init     sync.Once
	requests chan *Request[K, V]
	quit     chan struct{}
	closed   atomic.Bool
}

func (g *Group[K, V]) Do(ctx context.Context, key K, f func(context.Context, []*Request[K, V])) (V, error) {
	if g.closed.Load() {
		return *new(V), ErrClosed
	}
	g.init.Do(runLoop(g, f))

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	r := &Request[K, V]{
		Val:        key,
		C:          make(chan V, 1),
		CancelFunc: cancel,
		ctx:        ctx,
		c:          make(chan V, 1),
	}

	select {
	case g.requests <- r: // was able to put it on the batch queue
	case <-ctx.Done():
		return *new(V), context.Cause(ctx)
	}
	select {
	case res := <-r.c:
		return res, nil /* errors.New("<nil>") */
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
				callCtx, callCancel := context.WithCancel(context.Background())
				var cw atomic.Int64
				select {
				case r := <-g.requests:
					rr = append(rr, r)
					go waitCtx(r.ctx, callCancel, &cw, r.C, r.c)
				EMPTY:
					for {
						select {
						case r := <-g.requests:
							rr = append(rr, r)
							go waitCtx(r.ctx, callCancel, &cw, r.C, r.c)
						default:
							break EMPTY
						}
					}
					f(callCtx, rr)
					rr = rr[:0]
				case <-g.quit:
					callCancel()
					return
				}
			}
		}()
	}
}

func waitCtx[V any](ctx context.Context, cancel context.CancelFunc, cw *atomic.Int64, in <-chan V, out chan<- V) {
	cw.Add(1) // bump the counter
	var res V
	select {
	case <-ctx.Done():
	case res = <-in:
	}
	if cw.Add(-1) == 0 {
		cancel()
	}
	if err := ctx.Err(); err != nil {
		return
	}
	select {
	case <-ctx.Done():
	case out <- res:
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
	C          chan R
	c          chan R
	ctx        context.Context
	CancelFunc context.CancelCauseFunc
}

func (r Request[V, R]) Context() context.Context {
	if r.ctx == nil {
		return context.Background()
	}
	return r.ctx
}
