// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xhttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	. "go.adoublef.dev/net/xhttp"
	"go.adoublef.dev/testing/is"
)

func TestPingHandler(t *testing.T) {
	newGet := func(t testing.TB, p Pinger) func(context.Context) *http.Response {
		t.Helper()

		s := httptest.NewServer(PingHandler(p, 0))
		t.Cleanup(func() { s.Close() })

		get := func(ctx context.Context) *http.Response {
			t.Helper()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
			is.OK(t, err) // http.NewRequestWithContext

			res, err := s.Client().Do(req)
			is.OK(t, err) // http.Client.Do

			return res
		}
		return get
	}

	t.Run("OK", func(t *testing.T) {
		get := newGet(t, PingerFunc(func(ctx context.Context) error { return nil }))

		resp := get(t.Context())
		is.Equal(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("Sequential", func(t *testing.T) {
		var count int64
		get := newGet(t, PingerFunc(func(ctx context.Context) error { atomic.AddInt64(&count, 1); return nil }))

		for range 8 {
			_ = get(t.Context()).Body.Close()
		}

		is.Equal(t, count, 1)
	})

	t.Run("Concurrent", func(t *testing.T) {
		var count int64
		get := newGet(t, PingerFunc(func(ctx context.Context) error { atomic.AddInt64(&count, 1); return nil }))

		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() {
				_ = get(t.Context()).Body.Close()
			})
		}
		wg.Wait()

		is.Equal(t, count, 1)
	})
}
