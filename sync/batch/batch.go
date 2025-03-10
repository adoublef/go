// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package batch

import (
	"context"
	"slices"
	"sync"
	"time"

	"go.adoublef.dev/runtime/debug"
)

type Batch[K comparable, V any] struct {
	Data       K
	C          chan V
	CancelFunc context.CancelCauseFunc
	err        error // set if a user cancels
}

func (r *Batch[K, V]) Err() error { return r.err }

type Group[K comparable, V any] struct {
	BatchTimeout time.Duration
	wg           sync.WaitGroup
	init         sync.Once
	close        sync.Once
	calls        chan *Batch[K, V]
	quit         chan struct{}
}

func (g *Group[K, V]) Do(ctx context.Context, key K, f func(ctx context.Context, reqs ...Batch[K, V])) (V, error) {
	g.init.Do(func() {
		g.calls = make(chan *Batch[K, V], 16)
		g.quit = make(chan struct{})

		g.wg.Add(1)
		go g.runLoop(f)
	})

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	b := Batch[K, V]{
		Data:       key,
		C:          make(chan V, 1),
		CancelFunc: cancel,
	}

	select {
	case g.calls <- &b:
	case <-ctx.Done():
		return *new(V), ctx.Err()
	}

	select {
	case v := <-b.C:
		return v, nil
	case <-ctx.Done():
		b.err = context.Cause(ctx)
		return *new(V), b.err
	}
}

func (g *Group[K, V]) runLoop(f func(ctx context.Context, reqs ...Batch[K, V])) {
	defer g.wg.Done()
	calls := make([]Batch[K, V], 1<<10) // =1kb
	var idx int
	for {
		select {
		case c := <-g.calls:
			calls[idx] = *c
			idx++
		EMPTY:
			for {
				select {
				case c := <-g.calls:
					calls[idx] = *c
					idx++
				default:
					break EMPTY
				}
			}
			slices.Chunk(calls, idx)
			debug.Printf("sync/bacth: %d = idx", idx)
			switch d := g.BatchTimeout; {
			case d > 0:
				ctx, cancel := context.WithTimeout(context.Background(), d)
				f(ctx, calls[:idx]...)
				cancel()
			default:
				ctx, cancel := context.WithCancel(context.Background())
				f(ctx, calls[:idx]...)
				cancel()
			}
			//calls = calls[:0]
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
