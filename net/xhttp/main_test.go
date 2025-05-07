// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xhttp_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.adoublef.dev/net/nettest"
	. "go.adoublef.dev/net/xhttp"
)

// TestClient is configured for use within tests.
type TestClient struct {
	*http.Client
	*nettest.Proxy
	testing.TB
}

// Do sends an HTTP request and returns an HTTP response. The pattern follows similar rules to [http.ServeMux] in Go1.23.
// Options can be applied to modify the [http.Request] before sending it.
func (tc *TestClient) Do(ctx context.Context, pattern string, body io.Reader, opts ...func(*http.Request)) (*http.Response, error) {
	method, _, path, err := ParsePattern(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pattern: %v", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to return request: %v", err)
	}
	tc.Logf(`req, err := http.NewRequestWithContext(ctx, %q, %q, body)`, method, path)
	for _, o := range opts {
		o(req)
	}
	res, err := tc.Client.Do(req)
	tc.Logf(`res, %v := tc.Client.Do(req)`, err)
	return res, err
}

// newTestClient returns a new [TestClient] with the [httptest.Server] setup to near
// production levels. This server sits behind a proxy that can simulate network failures.
func newTestClient(tb testing.TB, h http.Handler) *TestClient {
	tb.Helper()

	ts := httptest.NewUnstartedServer(h)
	ts.Config.MaxHeaderBytes = DefaultMaxHeaderBytes
	// note: the client panics if readTimeout is less than the test timeout
	// is this a non-issue?
	ts.Config.ReadTimeout = DefaultReadTimeout
	ts.Config.WriteTimeout = DefaultWriteTimeout
	ts.Config.IdleTimeout = DefaultIdleTimeout
	// CipherSuites is a list of enabled TLS 1.0–1.2 cipher suites.
	// The order of the list is ignored.
	// Note that TLS 1.3 ciphersuites are not configurable.
	// ts.Config.TLSConfig.CipherSuites
	ts.StartTLS()

	proxy := nettest.NewProxy("HTTP_"+tb.Name(), strings.TrimPrefix(ts.URL, "https://"))
	if tp, ok := ts.Client().Transport.(*http.Transport); ok {
		tp.DisableCompression = true
	}
	tc := WithTransport(ts.Client(), "https://"+proxy.Listen())
	return &TestClient{tc, proxy, tb}
}
