// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package is provides a lightweight extension to the standard library's testing capabilities.
package is

import (
	"testing"

	"github.com/matryer/is"
)

// OK asserts that error is nil
func OK(tb testing.TB, err error) {
	is := is.NewRelaxed(tb)
	is.Helper()
	is.NoErr(err)
}

// Equal fails test if two values are not Equal
func Equal[V comparable](tb testing.TB, a, b V) {
	is := is.NewRelaxed(tb)
	is.Helper()
	is.Equal(a, b)
}

// True fails test if expression is false
func True(tb testing.TB, exp bool) {
	is := is.NewRelaxed(tb)
	is.Helper()
	is.True(exp)
}
