// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitset_test

import (
	"io"
	"testing"

	. "go.adoublef.dev/container/bitset"
)

var M = 100000

func TestBitUint8(t *testing.T) {
	t.Parallel()

	t.Run("Len", func(t *testing.T) {
		t.Parallel()
		b := NewBitUint8(M) // a bit set of 1
		if got, want := b.Len(), M; got != want {
			t.Errorf("BitUint8.Len: got=%d;want=%d", got, want)
		}
	})

	t.Run("ReadFrom", func(t *testing.T) {
		t.Parallel()

		a := NewBitUint8(M) // a bit set of 1
		for j := 2; j < M; j += 13 {
			a.Set(j, true)
		}
		for j := 1; j < M; j += 5 {
			a.Set(j, false)
		}

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

		var b BitUint8
		_, err := b.ReadFrom(pr)
		if err != nil {
			t.Errorf("BitUint8.ReadFrom: %v", err)
		}

		if got, want := b.Len(), a.Len(); got != want {
			t.Errorf("BitUint8.Len: got=%d;want=%d", got, want)
		}

		// check equality
		for j, n := 0, max(a.Len(), b.Len()); j < n; j++ {
			if a.Has(j) != b.Has(j) {
				t.Errorf("bitset %d differs at index %d", 1, j)
				return
			}
		}
	})
}

func TestBitBool(t *testing.T) {
	t.Parallel()

	t.Run("Len", func(t *testing.T) {
		t.Parallel()

		b := new(BitBool) // a bit set of 1
		for j := 2; j < M; j += 1 {
			b.Set(j, true)
		}
		if got, want := b.Len(), M; got != want {
			t.Errorf("got != want: got=%d;want=%d", got, want)
		}
	})
}
