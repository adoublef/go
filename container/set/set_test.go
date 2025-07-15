// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package set_test

import (
	"fmt"
	"slices"

	. "go.adoublef.dev/container/set"
)

func ExampleUnique() {
	s := []int{3, 1, 2, 0, 1, 3, 1, 4, 1, 3}
	for v := range Unique[int, TreeSet[int]](slices.Values(s)) {
		fmt.Println(v)
	}
	// Output:
	// 3
	// 1
	// 2
	// 0
	// 4
}

func ExampleTreeSet() {
	s := []int{3, 1, 2, 0, 1, 3, 1, 4, 1, 3}
	ts := new(TreeSet[int])
	Add(ts, slices.Values(s))
	for v := range ts.All() {
		fmt.Println(v)
	}
	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
}
