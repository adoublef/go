// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chans

import (
	"iter"
	"sync"
)

// OutFunc fan-out
func OutFunc[V any](f func(v V), seq iter.Seq[V]) {
	var wg sync.WaitGroup
	for v := range seq {
		wg.Add(1)
		go func() {
			f(v)
			wg.Done()
		}()
	}
	wg.Wait()
}

func OutFunc2[K, V any](f func(k K, v V), seq iter.Seq2[K, V]) {
	var wg sync.WaitGroup
	for k, v := range seq {
		wg.Add(1)
		go func() {
			f(k, v)
			wg.Done()
		}()
	}
	wg.Wait()
}

// OutChan fan-out with a channel
func OutChan[V, R any](f func(v V), out chan<- R, seq iter.Seq[V]) {
	var wg sync.WaitGroup
	for v := range seq {
		wg.Add(1)
		go func() {
			f(v)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
}

// OutChan fan-out with a channel
func OutChan2[T1, T2, R any](f func(k T1, v T2), out chan<- R, seq iter.Seq2[T1, T2]) {
	var wg sync.WaitGroup
	for k, v := range seq {
		wg.Add(1)
		go func() {
			f(k, v)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
}
