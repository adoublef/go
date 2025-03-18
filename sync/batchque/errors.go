// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package batchque

import (
	"errors"
	"sync"
)

func CancelFunc[K comparable, V any](err error, rr []Request[K, V]) {
	var wg sync.WaitGroup
	wg.Add(len(rr))
	for _, a := range rr {
		go func() {
			wg.Done()
			a.CancelFunc(err)
		}()
	}
	wg.Wait()
}

var ErrClosed = errors.New("use of closed connection")
