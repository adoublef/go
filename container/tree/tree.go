// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Pacakge tree
//
// See blog: https://go.dev/blog/generic-interfaces
package tree

import (
	"cmp"
	"iter"
)

type OrderedComparer[E any] interface {
	comparable // cmp.Ordered
	Comparer[E]
}

type Comparer[T any] interface {
	Compare(T) int
}

type node[E any] struct {
	value E
	left  *node[E]
	right *node[E]
}

func (n *node[E]) add(cmp func(E, E) int, element E) *node[E] {
	if n == nil {
		return &node[E]{value: element}
	}
	switch sign := cmp(element, n.value); {
	case sign < 0:
		n.left = n.left.add(cmp, element)
	case sign > 0:
		n.right = n.right.add(cmp, element)
	}
	return n
}

func (n *node[E]) has(cmp func(E, E) int, element E) bool {
	if n == nil {
		return false
	}
	switch sign := cmp(element, n.value); {
	case sign < 0:
		return n.left.has(cmp, element)
	case sign > 0:
		return n.right.has(cmp, element)
	default:
		return true
	}
}

func (n *node[E]) all(yield func(E) bool) bool {
	return n == nil || (n.left.all(yield) && yield(n.value) && n.right.all(yield))
}

// The zero value of a Tree is a ready-to-use empty tree.
//
//	var t Tree[string]
//	t.Insert("Garry Kasparov")
//	t.Insert("Magnus Carlsen")
//	t.Insert("Bobby Fischer")
//	t.Insert("Anatoly Karpov")
//	t.Insert("Mikhail Tal")
//	for n := range t.All() {
//	    fmt.Println(n)
//	}
type Tree[E cmp.Ordered] struct {
	root *node[E]
}

func (t *Tree[E]) Add(element E) {
	t.root = t.root.add(cmp.Compare[E], element)
}

func (t *Tree[E]) Has(element E) bool {
	return t.root.has(cmp.Compare[E], element)
}

func (t *Tree[E]) All() iter.Seq[E] {
	return func(yield func(E) bool) {
		t.root.all(yield)
	}
}

// The zero value of a MethodTree is a ready-to-use empty tree.
//
//	type Player struct {
//	    Name   string
//	    Rating int
//	}
//	func (p Player) Compare(q Player) int {
//	     return cmp.Compare(p.Rating, q.Rating)
//	}
//
//	var t MethodTree[Player]
//	t.Insert(Player{"Garry Kasparov", 2851})
//	t.Insert(Player{"Magnus Carlsen", 2882})
//	t.Insert(Player{"Bobby Fischer", 2785})
//	t.Insert(Player{"Anatoly Karpov", 2780})
//	t.Insert(Player{"Mikhail Tal", 2705})
//
//	for p := range t.All() {
//	     fmt.Println(p)
//	}
type MethodTree[E Comparer[E]] struct {
	root *node[E]
}

func (t *MethodTree[E]) Add(e E) {
	t.root = t.root.add(E.Compare, e)
}

func (t *MethodTree[E]) Has(element E) bool {
	return t.root.has(E.Compare, element)
}

func (t *MethodTree[E]) All() iter.Seq[E] {
	return func(yield func(E) bool) {
		t.root.all(yield)
	}
}

// A FuncTree must be created with NewTreeFunc.
//
//	type Player struct {
//	    Name   string
//	    Rating int
//	}
//
//	func (p Player) Compare(q Player) int {
//	    return cmp.Compare(p.Rating, q.Rating)
//	}
//
//	players := []Player{
//	    {"Garry Kasparov", 2851},
//	    {"Magnus Carlsen", 2882},
//	    {"Bobby Fischer", 2785},
//	    {"Anatoly Karpov", 2780},
//	    {"Mikhail Tal", 2705},
//	}
//
//	t := NewFuncTree(func(a, b Player) int {
//	    return cmp.Compare(a.Rating, b.Rating)
//	})
//
//	t.Insert(Player{"Garry Kasparov", 2851})
//	t.Insert(Player{"Magnus Carlsen", 2882})
//	t.Insert(Player{"Bobby Fischer", 2785})
//	t.Insert(Player{"Anatoly Karpov", 2780})
//	t.Insert(Player{"Mikhail Tal", 2705})
//
//	for p := range t.All() {
//	    fmt.Println(p)
//	}
type FuncTree[E any] struct {
	root *node[E]
	cmp  func(E, E) int
}

func NewFuncTree[E any](cmp func(E, E) int) *FuncTree[E] {
	return &FuncTree[E]{cmp: cmp}
}

func (t *FuncTree[E]) Add(element E) {
	t.root = t.root.add(t.cmp, element)
}

func (t *FuncTree[E]) Has(element E) bool {
	return t.root.has(t.cmp, element)
}

func (t *FuncTree[E]) All() iter.Seq[E] {
	return func(yield func(E) bool) {
		t.root.all(yield)
	}
}
