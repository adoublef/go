// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package set
//
// See more: https://go.dev/blog/generic-interfacess
package set

import (
	"cmp"
	"iter"
	"maps"

	"go.adoublef.dev/container/tree"
)

type Set[E any] interface {
	Add(E)
	Has(E) bool
	All() iter.Seq[E]
}

// Add adds all unique elements from seq into set.
func Add[E any](set Set[E], seq iter.Seq[E]) {
	for v := range seq {
		set.Add(v)
	}
}

// Unique removes duplicate elements from the input sequence, yielding only
// the first instance of any element.
func Unique[E, S any, P interface {
	*S
	Set[E]
}](input iter.Seq[E]) iter.Seq[E] {
	return func(yield func(E) bool) {
		// We convert to PS, as only that is constrained to have the methods.
		// The conversion is allowed, because the type set of PS only contains *S.
		seen := P(new(S))
		for v := range input {
			if seen.Has(v) {
				continue
			}
			if !yield(v) {
				return
			}
			seen.Add(v)
		}
	}
}

type TreeSet[E cmp.Ordered] struct {
	root     tree.Tree[E]
	elements map[E]struct{}
}

func (s *TreeSet[E]) Add(element E) {
	if s.elements == nil {
		s.elements = make(map[E]struct{})
	}
	if _, ok := s.elements[element]; ok {
		return
	}
	s.elements[element] = struct{}{}
	s.root.Add(element)
}

func (s *TreeSet[E]) Has(e E) bool { _, ok := s.elements[e]; return ok }

func (s *TreeSet[E]) All() iter.Seq[E] { return s.root.All() }

// OrderedSet
//
//	type Player struct {
//	    Name   string
//	    Rating int
//	}
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
//	var s OrderedSet[Player]
//	for _, p := range players {
//	     s.Add(p)
//	}
//	for p := range s.All() {
//	    fmt.Println(p)
//	}
type OrderedSet[E tree.OrderedComparer[E]] struct {
	tree     tree.MethodTree[E] // for efficient iteration in order
	elements map[E]struct{}     // for (near) constant time lookup
}

func (s *OrderedSet[E]) Add(e E) {
	if s.elements == nil {
		s.elements = make(map[E]struct{})
	}
	if _, ok := s.elements[e]; ok {
		return
	}
	s.elements[e] = struct{}{}
	s.tree.Add(e)
}

func (s *OrderedSet[E]) Has(e E) bool { _, ok := s.elements[e]; return ok }

func (s *OrderedSet[E]) All() iter.Seq[E] { return s.tree.All() }

type HashSet[E comparable] map[E]struct{}

func (s HashSet[E]) Add(v E)          { s[v] = struct{}{} }
func (s HashSet[E]) Delete(v E)       { delete(s, v) }
func (s HashSet[E]) Has(v E) bool     { _, ok := s[v]; return ok }
func (s HashSet[E]) All() iter.Seq[E] { return maps.Keys(s) }
