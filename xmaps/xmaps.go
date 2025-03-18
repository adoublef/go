// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package xmaps defines various functions useful with maps of any type.
package xmaps

import (
	"slices"

	"go.adoublef.dev/xiter"
)

// GroupFunc groups slice elements into a map of slices using a key function.
func GroupFunc[K comparable, Slice ~[]V, V any](aa Slice, f func(V) K) map[K]Slice {
	return xiter.Reduce(func(m map[K]Slice, a V) map[K]Slice {
		k := f(a)
		m[k] = append(m[k], a)
		return m
	}, make(map[K]Slice), slices.Values(aa))
}

// SetFunc creates a set (as a map to empty structs) from a slice using a key function.
func SetFunc[K comparable, Slice ~[]V, V any](aa Slice, f func(V) K) map[K]struct{} {
	return xiter.Reduce(func(m map[K]struct{}, a V) map[K]struct{} {
		k := f(a)
		m[k] = struct{}{}
		return m
	}, make(map[K]struct{}), slices.Values(aa))
}
