// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hashque_test

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	. "go.adoublef.dev/sync/hashque"
	"go.adoublef.dev/testing/is"
)

func TestGroup(t *testing.T) {
	t.Run("Do", func(t *testing.T) {
		const Delta = 10000

		var g Group[string]
		var count int

		var wg sync.WaitGroup
		wg.Add(Delta)

		for i := range Delta {
			go func() {
				g.Do("1", func() {
					// wait group needed next to the work
					count += (i + 1)
					wg.Done()
				})
			}()
		}
		wg.Wait()

		is.Equal(t, count, (Delta*(Delta+1))/2) // sum(100)
	})

	t.Run("Keyed", func(t *testing.T) {
		const Delta = 10000

		var g Group[string]
		var odd, even int

		var wg sync.WaitGroup
		wg.Add(Delta)

		for i := range Delta {
			go func() {
				g.Do(strconv.Itoa((i%2)+1), func() {
					if i%2 == 0 {
						odd += (i + 1)
					} else {
						even += (i + 1)
					}
					wg.Done()
				})
			}()
		}
		wg.Wait()

		is.Equal(t, odd, (Delta/2)*(Delta/2))
		is.Equal(t, even, (Delta/2)*((Delta/2)+1))
	})

	t.Run("TryDo", func(t *testing.T) {
		const Delta = 100 // a message should fail with this load

		var g Group[string]
		var success, failed atomic.Int64

		var wg sync.WaitGroup
		wg.Add(Delta)

		for range Delta {
			go func() {
				if ok := g.TryDo("1", func() {
					success.Add(1)
					wg.Done()
				}); !ok {
					failed.Add(1)
					wg.Done()
				}
			}()
		}
		wg.Wait()

		ns, nf := success.Load(), failed.Load()
		is.True(t, nf > 0)
		is.Equal(t, ns+nf, int64(Delta))
	})
}
