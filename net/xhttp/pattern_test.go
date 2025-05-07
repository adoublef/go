// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xhttp_test

import (
	"testing"

	. "go.adoublef.dev/net/xhttp"
)

func Test_ParsePattern(t *testing.T) {
	type testcase struct {
		pattern      string
		method, path string
	}

	tt := []testcase{
		{"GET /path", "GET", "/path"},
		{"GET", "", ""},
	}

	for _, tc := range tt {
		t.Run(tc.pattern, func(t *testing.T) {
			method, _, path, _ := ParsePattern(tc.pattern)
			if got, want := method, tc.method; got != want {
				t.Errorf("unexpected method value: got %v; want %v", got, want)
			}
			if got, want := path, tc.path; got != want {
				t.Errorf("unexpected path value: got %v; want %v", got, want)
			}
		})
	}
}
