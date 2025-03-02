// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package du_test

import (
	"testing"

	. "go.adoublef.dev/os/du"
)

func TestParseSize(t *testing.T) {
	type testcase struct {
		s       string
		wantErr bool
	}

	tt := []testcase{
		{"", false},
		{"1", false},
		{"1g", false},
		{"1G", false},
		{"1gb", false},
		{"1GB", false},
		{"1.0GB", false},
		{"1 GB", true},
	}

	for _, tc := range tt {
		t.Run(tc.s, func(t *testing.T) {
			_, err := ParseSize(tc.s)
			if (err == nil) == tc.wantErr {
				t.Errorf("unexpected error: got %v", err)
			}
		})
	}
}
