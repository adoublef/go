// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package batchque

import "sync"

func CancelFunc[K comparable, V any](rr []*Request[K, V], err error) {
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
