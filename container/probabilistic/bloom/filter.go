// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bloom

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"go.adoublef.dev/container/set"
)

type Filter struct {
	Hasher Hasher
	m      int // m size of bitset
	k      int // k number of sets
	set    set.BitUint8
}

func (f *Filter) Set(v string) {
	// Double Hashing
	h := f.Hasher.Hash([]byte(v))
	u := uint32(h /* & 0xffffffff */)
	l := uint32((h >> 32) /* & 0xffffffff */)
	for i := 0; i < f.k; i++ {
		h := (l + u*uint32(i)) % uint32(f.m)
		f.set.Set(int(h), true)
	}
}

func (f *Filter) Has(v string) bool {
	// Double Hashing
	h := f.Hasher.Hash([]byte(v))
	u := uint32(h /* & 0xffffffff */)
	l := uint32((h >> 32) /* & 0xffffffff */)
	for i := 0; i < f.k; i++ {
		h := (l + u*uint32(i)) % uint32(f.m)
		if !f.set.Has(int(h)) {
			return false
		}
	}
	return true
}

func (f Filter) WriteTo(w io.Writer) (n int64, err error) {
	m := int64(f.m)
	err = binary.Write(w, binary.LittleEndian, &m)
	if err != nil {
		return n, fmt.Errorf("cannot write size of bitset: %w", err)
	}
	n += 8

	k := int64(f.k)
	err = binary.Write(w, binary.LittleEndian, &k)
	if err != nil {
		return n, fmt.Errorf("cannot write number of hashes: %w", err)
	}
	n += 8

	nw, err := f.set.WriteTo(w)
	n += nw
	return n, err
}

func (f *Filter) ReadFrom(r io.Reader) (n int64, err error) {
	var m int64
	err = binary.Read(r, binary.LittleEndian, &m)
	if err != nil {
		return 0, fmt.Errorf("cannot read size of bitset: %w", err)
	}
	n += 8
	f.m = int(m)

	var k int64
	err = binary.Read(r, binary.LittleEndian, &k)
	if err != nil {
		return n, fmt.Errorf("cannot read number of hashes: %w", err)
	}
	n += 8
	f.k = int(k)

	nr, err := f.set.ReadFrom(r)
	n += nr
	return n, err
}

func NewFilter(n int, p float64, hf Hasher) *Filter {
	assert(n > 0, "n must be positive")
	assert(p > 0 && p < 1, "p must be exclusively between 0 and 1")
	assert(hf != nil, "hasher cannot be nil")

	m := math.Ceil((float64(n) * math.Log(p)) / math.Log(1/math.Pow(2, math.Log(2))))
	k := math.Round((m / float64(n)) * math.Log(2))
	bs := set.NewBitUint8(int(m * k)) //make([]uint8, (m*k+7)/8)
	return &Filter{m: int(m), k: int(k), set: bs, Hasher: hf}
}

type Hasher interface {
	Hash(b []byte) uint64
}

type HashFunc func(b []byte) uint64

func (hf HashFunc) Hash(b []byte) uint64 {
	return hf(b)
}

func assert(exp bool, format string) {
	if !exp {
		panic(format)
	}
}
