// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bloom_test

import (
	"io"
	"strconv"
	"testing"

	. "go.adoublef.dev/container/probabilistic/bloom"
)

func TestFilter(t *testing.T) {
	t.Parallel()
	hf := HashFunc(func(key []byte) uint64 {
		i, err := strconv.Atoi(string(key))
		if err != nil {
			panic(err)
		}
		return uint64(i)
	})

	t.Run("Has", func(t *testing.T) {
		t.Parallel()

		f := NewFilter(10, 0.01, hf)

		f.Set("1")
		f.Set("2")
		f.Set("3")

		if got := f.Has("3"); !got {
			t.Errorf("Has(%s) want %t, got %t", "3", true, got)
		}
		if got := f.Has("5"); got {
			t.Errorf("Has(%s) want %t, got %t", "5", false, got)
		}
	})
	t.Run("ReadFrom", func(t *testing.T) {
		t.Parallel()

		a := NewFilter(10, 0.01, hf)

		a.Set("1")
		a.Set("2")
		a.Set("3")

		pr, pw := io.Pipe()
		defer pr.Close()
		go func() {
			defer pw.Close()
			_, err := a.WriteTo(pw)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
		}()

		var b Filter
		b.Hasher = hf
		_, err := b.ReadFrom(pr)
		if err != nil {
			t.Errorf("BitUint8.ReadFrom: %v", err)
		}

		if got := b.Has("3"); !got {
			t.Errorf("Has(%s) want %t, got %t", "3", true, got)
		}
		if got := b.Has("5"); got {
			t.Errorf("Has(%s) want %t, got %t", "5", false, got)
		}
	})
}
