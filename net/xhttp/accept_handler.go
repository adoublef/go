// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xhttp

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"

	"github.com/golang/gddo/httputil"
	"go.adoublef.dev/runtime/debug"
)

var (
	ContentTypOfferKey = &contextKey{"accept-offer"}
)

// AcceptHandler verifies the client can accept the response of a request.
func AcceptHandler(h http.Handler /* custom types */) http.Handler {
	var (
		ct = []string{"application/json", "text/html"}
		ce = []string{"identity", "gzip" /* "deflate", "zstd", "zlib" */}
	)
	// note: if we allow compression option
	// panic if invalid
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		accept := httputil.NegotiateContentType(r, ct, "")
		if accept == "" {
			http.Error(w, `Only "application/json" or "text/html" content types supported`, http.StatusNotAcceptable)
			return
		}
		debug.Printf("AcceptHandler: %q = httputil.NegotiateContentType(r, ct, _)", accept)

		ctx = context.WithValue(ctx, ContentTypOfferKey, accept)
		encoding := httputil.NegotiateContentEncoding(r, ce)
		// note: should we always encode text/html?
		if encoding == "" || encoding == ce[0] {
			h.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		debug.Printf("AcceptHandler: %q = negotiate.ContentEncoding(r, ce)", encoding)

		w.Header().Set("Content-Encoding", encoding)
		// todo: support other encoding types?
		gw, _ := gzip.NewWriterLevel(w, gzip.DefaultCompression)
		defer gw.Close() // should I defer?
		h.ServeHTTP(&gzipWriter{w, gw}, r.WithContext(ctx))
	})
}

type gzipWriter struct {
	http.ResponseWriter
	io.Writer
}

// Write implements http.ResponseWriter.
func (w *gzipWriter) Write(p []byte) (int, error) {
	return w.Writer.Write(p)
}
