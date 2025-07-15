// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tree_test

import (
	"cmp"
	"fmt"

	. "go.adoublef.dev/container/tree"
)

func ExampleTree() {
	var t Tree[string]
	t.Add("Garry Kasparov")
	t.Add("Magnus Carlsen")
	t.Add("Bobby Fischer")
	t.Add("Anatoly Karpov")
	t.Add("Mikhail Tal")
	for n := range t.All() {
		fmt.Println(n)
	}
	// Output:
	// Anatoly Karpov
	// Bobby Fischer
	// Garry Kasparov
	// Magnus Carlsen
	// Mikhail Tal
}

type Player struct {
	Name   string
	Rating int
}

func (p Player) Compare(q Player) int {
	return cmp.Compare(p.Rating, q.Rating)
}

func ExampleMethodTree() {
	var t MethodTree[Player]
	t.Add(Player{"Garry Kasparov", 2851})
	t.Add(Player{"Magnus Carlsen", 2882})
	t.Add(Player{"Bobby Fischer", 2785})
	t.Add(Player{"Anatoly Karpov", 2780})
	t.Add(Player{"Mikhail Tal", 2705})
	for n := range t.All() {
		fmt.Println(n)
	}
	// Output:
	// {Mikhail Tal 2705}
	// {Anatoly Karpov 2780}
	// {Bobby Fischer 2785}
	// {Garry Kasparov 2851}
	// {Magnus Carlsen 2882}
}

func ExampleFuncTree() {
	t := NewFuncTree(func(a, b Player) int {
		return cmp.Compare(a.Rating, b.Rating)
	})
	t.Add(Player{"Garry Kasparov", 2851})
	t.Add(Player{"Magnus Carlsen", 2882})
	t.Add(Player{"Bobby Fischer", 2785})
	t.Add(Player{"Anatoly Karpov", 2780})
	t.Add(Player{"Mikhail Tal", 2705})
	for n := range t.All() {
		fmt.Println(n)
	}
	// Output:
	// {Mikhail Tal 2705}
	// {Anatoly Karpov 2780}
	// {Bobby Fischer 2785}
	// {Garry Kasparov 2851}
	// {Magnus Carlsen 2882}
}
