// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xhttp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "go.adoublef.dev/net/xhttp"
	"go.adoublef.dev/testing/is"
)

func Test_MethodOverride(t *testing.T) {
	t.Run("Query", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /{$}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		s := httptest.NewServer(MethodOverride(mux))
		t.Cleanup(func() { s.Close() })

		rs, err := s.Client().Post(s.URL+"?_method=DELETE", "", nil)
		is.OK(t, err) // (http.Client).Post
		is.Equal(t, rs.StatusCode, http.StatusOK)
	})

	t.Run("Header", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /{$}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		s := httptest.NewServer(MethodOverride(mux))
		t.Cleanup(func() { s.Close() })

		r, err := http.NewRequest(http.MethodPost, s.URL, nil)
		is.OK(t, err) // http.NewRequest

		r.Header.Set("X-HTTP-Method-Override", http.MethodDelete)

		rs, err := s.Client().Do(r)
		is.OK(t, err) // (http.Client).Post
		is.Equal(t, rs.StatusCode, http.StatusOK)
	})
}
