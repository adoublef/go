// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package batch_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	. "go.adoublef.dev/sync/batch"
	"go.adoublef.dev/testing/is"
)

func TestGroup_Do(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var g Group[string, int]

		ctx := t.Context()
		t.Cleanup(func() {
			g.Close() // should close
		})

		// Make concurrent requests
		var wg sync.WaitGroup
		results := make([]int, 5)
		errs := make([]error, 5)

		testInputs := []string{"one", "two", "three", "four", "five"}

		for i, input := range testInputs {
			wg.Add(1)
			go func(idx int, str string) {
				defer wg.Done()
				res, err := g.Do(ctx, str, func(_ context.Context, reqs ...Batch[string, int]) {
					for _, r := range reqs {
						if n := len(r.Data); n > 10 {
							r.CancelFunc(errors.New("too many characters"))
						} else {
							r.C <- n
						}
					}
				})
				results[idx] = res
				errs[idx] = err
			}(i, input)
		}

		wg.Wait()

		// Verify all concurrent results
		for i, input := range testInputs {
			is.OK(t, errs[i])
			is.Equal(t, results[i], len(input))
		}

		// g.Close() // should close
	})

	t.Run("Large", func(t *testing.T) {
		var g Group[int, int]
		ctx := context.Background()
		t.Cleanup(func() {
			g.Close() // should close
		})

		// Launch 2000 concurrent requests (more than fixed array size)
		var wg sync.WaitGroup
		for i := range 2000 {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				_, err := g.Do(ctx, i, func(ctx context.Context, messages ...Batch[int, int]) {
					for _, msg := range messages {
						msg.C <- msg.Data
					}
				})
				is.OK(t, err) // Should not panic or error
			}(i)
		}
		wg.Wait()
		// g.Close()
	})

	t.Run("CancelFunc", func(t *testing.T) {
		var g Group[int, int]
		ctx := context.Background()
		t.Cleanup(func() {
			g.Close() // should close
		})

		// Test data: even numbers will succeed, odd numbers will error
		const N = 100
		rs := make([]int, N)
		errs := make([]error, N)

		var wg sync.WaitGroup
		for i := range N {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				r, err := g.Do(ctx, idx, func(ctx context.Context, reqs ...Batch[int, int]) {
					for _, r := range reqs {
						if r.Data%2 == 0 {
							// Even numbers succeed
							r.C <- r.Data * 2
						} else {
							// Odd numbers return an error
							errMsg := fmt.Sprintf("error processing odd number: %d", r.Data)
							r.CancelFunc(errors.New(errMsg))
						}
					}
				})
				rs[idx] = r
				errs[idx] = err
			}(i)
		}
		wg.Wait()

		// Verify results
		for i := range N {
			if i%2 == 0 {
				// Even numbers should succeed with doubled value
				is.OK(t, errs[i])
				is.Equal(t, rs[i], i*2)
			} else {
				// Odd numbers should return an error
				is.True(t, errs[i] != nil)
				want := fmt.Sprintf("error processing odd number: %d", i)
				is.True(t, strings.Contains(errs[i].Error(), want))
			}
		}

		// g.Close()
	})
}
