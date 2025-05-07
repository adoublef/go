// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xhttp_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Shopify/toxiproxy/v2/toxics"
	. "go.adoublef.dev/net/xhttp"
	"go.adoublef.dev/testing/is"
)

func Test_Decode(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		tc, ctx := newDecodeClient(t, DefaultMaxBytes, 0), context.Background()

		s := `{"username":"username","password":"password"}`
		res, err := tc.Do(ctx, "POST /", strings.NewReader(s), ctJSON)
		is.OK(t, err)
		is.Equal(t, res.StatusCode, http.StatusOK)
	})

	t.Run("ErrRequestEntityTooLarge", func(t *testing.T) {
		tc, ctx := newDecodeClient(t, 1, 0), context.Background()

		s := `{"username":"username","password":"password"}`
		res, err := tc.Do(ctx, "POST /", strings.NewReader(s), ctJSON)
		is.OK(t, err)
		is.Equal(t, res.StatusCode, http.StatusRequestEntityTooLarge)
	})

	t.Run("ErrRequestTimeout", func(t *testing.T) {
		tc, ctx := newDecodeClient(t, DefaultMaxBytes, 1), context.Background()

		toxic, err := tc.AddToxic("bandwidth", true, &toxics.BandwidthToxic{Rate: 1})
		is.OK(t, err)
		t.Logf("%v := tc.AddToxic(bandwidth)", err)

		s := fmt.Sprintf(`{"username":%q,"password":"password"}`, strings.Repeat("username", 1<<10))
		res, err := tc.Do(ctx, "POST /", strings.NewReader(s), ctJSON)
		is.OK(t, err)
		is.Equal(t, res.StatusCode, http.StatusRequestTimeout)

		is.OK(t, tc.RemoveToxic(toxic))
	})

	t.Run("ErrBadRequest", func(t *testing.T) {
		c, ctx := newDecodeClient(t, DefaultMaxBytes, 0), context.Background()

		type testcase struct {
			body   string
			detail string
		}

		for name, tc := range map[string]testcase{
			"Syntax": {
				body:   `{"username:"user"}`,
				detail: "invalid character 'u' at position 13",
			},
			"Syntax2": {
				body:   `<"username:"user"}`,
				detail: "invalid character '<' at position 1",
			},
			"Unmarshal": {
				body:   `{"username":1,"password":"pass"}`,
				detail: `unexpected number for field "username" at position 13`,
			},
			"Unmarshal2": {
				body:   `"username:"user"}`,
				detail: "unexpected string for field \"\" at position 11",
				// skip:   true,
			},
			"UnknownField": {
				body:   `{"never":"user"}`,
				detail: `unknown field "never"`,
			},
			"Stream": {
				body: `{"username":"username","password":"password"}{}`,
			},
		} {
			t.Run(name, func(t *testing.T) {
				// create digest
				res, err := c.Do(ctx, "POST /", strings.NewReader(tc.body), ctJSON)
				is.OK(t, err)
				is.Equal(t, res.StatusCode, http.StatusBadRequest)
				// todo: read the error to see if it matches
			})
		}
	})
}

var ctJSON = func(r *http.Request) { r.Header.Set("Content-Type", "application/json") }

func newDecodeClient(tb testing.TB, sz int, d time.Duration) *TestClient {
	handleTest := func() http.HandlerFunc {
		type payload struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		return func(w http.ResponseWriter, r *http.Request) {
			_, err := Decode[payload](w, r, sz, d)
			if err != nil {
				var de *DecodeError
				if errors.As(err, &de) {
					w.WriteHeader(de.Code)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /{$}", handleTest())

	return newTestClient(tb, mux)
}
