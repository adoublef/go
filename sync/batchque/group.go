// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package batchque

import (
	"context"
	"sync"
	"sync/atomic"

	"go.adoublef.dev/runtime/debug"
)

type Group[In, Out any] struct {
	wg       sync.WaitGroup
	init     sync.Once
	requests chan Request[In, Out]
	quit     chan struct{}
	closed   atomic.Bool
}

func (g *Group[In, Out]) Do(ctx context.Context, key In, f func(context.Context, []Request[In, Out])) (Out, error) {
	if g.closed.Load() {
		return *new(Out), ErrClosed
	}
	g.init.Do(runLoop(g, f))

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	c := make(chan Out, 1)
	r := Request[In, Out]{
		Val:        key,
		C:          c,
		CancelFunc: cancel,
		ctx:        ctx,
	}

	select {
	case g.requests <- r: // was able to put it on the batch queue
	case <-ctx.Done():
		return *new(Out), context.Cause(ctx)
	}
	select {
	case res := <-c:
		return res, nil
	case <-ctx.Done():
		return *new(Out), context.Cause(ctx)
	}
}

func runLoop[In, Out any](g *Group[In, Out], f func(context.Context, []Request[In, Out])) func() {
	merge := func(ss []func() bool, ctx context.Context, cancel context.CancelFunc, n *atomic.Int64) []func() bool {
		n.Add(1)
		return append(ss, context.AfterFunc(ctx, func() {
			if n.Add(-1) == 0 {
				cancel()
			}
		}))
	}

	return func() {
		g.requests = make(chan Request[In, Out], 16) // backpressure?
		g.quit = make(chan struct{})

		g.wg.Add(1)
		go func() {
			defer g.wg.Done()
			rr := make([]Request[In, Out], 0, 1<<10) // =1kb
			ss := make([]func() bool, 0, 1<<10)      // =1kb
			for {
				ctx, cancel := context.WithCancel(context.Background())
				var n atomic.Int64
				select {
				case r := <-g.requests:
					rr = append(rr, r)
					ss = merge(ss, r.ctx, cancel, &n)
				EMPTY:
					for {
						select {
						case r := <-g.requests:
							rr = append(rr, r)
							ss = merge(ss, r.ctx, cancel, &n)
						default:
							break EMPTY
						}
					}
					debug.Printf("sync/batchque: %d = len(rr)", len(rr))
					f(ctx, rr)
					for _, stop := range ss {
						stop()
					}
					cancel()
					rr = rr[:0]
					ss = ss[:0]
				case <-g.quit:
					cancel()
					return
				}
			}
		}()
	}
}

func (g *Group[In, Out]) Close() error {
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
	C          chan<- R // public
	ctx        context.Context
	CancelFunc context.CancelCauseFunc
}

func (r Request[V, R]) Context() context.Context {
	if r.ctx == nil {
		return context.Background()
	}
	return r.ctx
}
