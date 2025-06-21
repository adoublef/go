// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xhttp_test

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	. "go.adoublef.dev/net/xhttp"
	"go.adoublef.dev/testing/is"
)

const (
	contentTypJSON = "application/json"
	contentTypHTML = "text/html"
)

var acceptAll = func(r *http.Request) { r.Header.Set("Accept", "*/*") }

func Test_AcceptHandler(t *testing.T) {
	accept := func(s string) func(*http.Request) {
		return func(r *http.Request) { r.Header.Set("Accept", s) }
	}
	acceptAll := accept("*/*")
	acceptEnc := func(s string) func(*http.Request) {
		return func(r *http.Request) { r.Header.Set("Accept-Encoding", s) }
	}

	t.Run("ContentTyp", func(t *testing.T) {

		c, ctx := newAcceptClient(t), t.Context()

		type testcase struct {
			accept     string
			contentTyp string
		}

		for _, tc := range []testcase{
			{accept: "*/*", contentTyp: contentTypJSON},
			{accept: "text/html", contentTyp: contentTypHTML},
			{accept: "application/json", contentTyp: contentTypJSON},
		} {
			res, err := c.Do(ctx, "GET /", nil, accept(tc.accept))
			is.OK(t, err)
			is.Equal(t, res.StatusCode, http.StatusOK)                 // got;want statusCode
			is.Equal(t, res.Header.Get("Content-Type"), tc.contentTyp) // got;want contentType
		}
	})

	t.Run("ContentEncode", func(t *testing.T) {

		c, ctx := newAcceptClient(t), t.Context()

		type testcase struct {
			accept string
			encode string
		}

		for _, tc := range []testcase{
			{accept: "gzip", encode: "gzip"},
			{accept: "", encode: ""},
			{accept: "identity", encode: ""},
		} {
			res, err := c.Do(ctx, "GET /", nil, acceptAll, acceptEnc(tc.accept))
			is.OK(t, err)
			is.Equal(t, res.StatusCode, http.StatusOK)                 // got;want statusCode
			is.Equal(t, res.Header.Get("Content-Encoding"), tc.encode) // got;want contentEncoding
		}
	})

	t.Run("Gzip", func(t *testing.T) {
		c, ctx := newAcceptClient(t), t.Context()

		res, err := c.Do(ctx, "GET /", nil, accept("text/html"), acceptEnc("gzip"))
		is.OK(t, err)

		gr, err := gzip.NewReader(res.Body)
		is.OK(t, err)

		p, err := io.ReadAll(gr) // ��)�+I�(��(�ͱ�/����+d
		is.OK(t, err)
		is.OK(t, res.Body.Close())

		is.Equal(t, string(p), "<p>text/html</p>") // got;want body
	})

	t.Run("ErrNotSet", func(t *testing.T) {
		c, ctx := newAcceptClient(t), t.Context()

		res, err := c.Do(ctx, "GET /", nil)
		is.OK(t, err)
		is.Equal(t, res.StatusCode, http.StatusNotAcceptable) // got;want statusCode
	})

	t.Run("ErrNotValid", func(t *testing.T) {
		c, ctx := newAcceptClient(t), t.Context()

		res, err := c.Do(ctx, "GET /", nil, accept("application/xml"))
		is.OK(t, err)
		is.Equal(t, res.StatusCode, http.StatusNotAcceptable) // got;want statusCode
	})
}

func newAcceptClient(tb testing.TB) *TestClient {
	tb.Helper()
	// encode json data as a response
	handleTest := func() http.HandlerFunc {
		type body struct {
			Typ string `json:"contentType"`
		}

		return func(w http.ResponseWriter, r *http.Request) {
			// I could set this in the handler
			accept, ok := r.Context().Value(ContentTypOfferKey).(string)
			if !ok {
				// log error
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			tb.Logf("%q, ok := r.Context().Value(AcceptKey).(string)", accept)

			w.Header().Set("Content-Type", accept)
			if accept == contentTypJSON {
				err := json.NewEncoder(w).Encode(&body{accept})
				tb.Logf("%v := json.NewEncoder(w).Encode(_)", err)
				return
			}
			// return html
			_, err := fmt.Fprintf(w, "<p>%s</p>", accept)
			tb.Logf("%v := fmt.Fprintf(w, _, accept)", err)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", handleTest())
	return newTestClient(tb, AcceptHandler(mux))
}
