// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache_test

import (
	"testing"

	. "go.adoublef.dev/sync/cache"
)

func Test_Cache_Get(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		// 1.
		var c LRU[string, string]
		// 1. add the value
		c.Add("1", "Hello")
		// 2. get the value
		hello, ok := c.Get("1")
		if !ok {
			t.Errorf("using key %q should have a value", "1")
		}
		if hello != "Hello" {
			t.Error("unexpected error")
		}
		if c.Bytes() != 7 {
			t.Error("unexpected error")
		}
	})

	t.Run("Complex", func(t *testing.T) {
		type simple struct {
			I int64
			S string
		}

		// 1.
		var c LRU[string, simple]
		// 1. add the value
		_ = c.Add("1", simple{23, ""})
		// 2. get the value
		hello, ok := c.Get("1")
		if !ok {
			t.Errorf("using key %q should have a value", "1")
		}
		if (hello != simple{23, ""}) {
			t.Error("unexpected missing")
		}
		t.Logf("size: %d", c.Bytes()) // why is this 16 bytes?
		if c.Bytes() != 16 {
			t.Error("unexpected size")
		}
	})
}
