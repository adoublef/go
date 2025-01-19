// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package set

import (
	"encoding/binary"
	"fmt"
	"io"
)

type BitUint8 []uint8

func (b BitUint8) Has(n int) bool {
	pos := n / 8
	j := n % 8
	return (b[pos] & (uint8(1) << j)) != 0
}

func (b BitUint8) Set(n int, t bool) {
	pos := n / 8
	j := uint(n % 8)
	if t {
		b[pos] |= (uint8(1) << j)
	} else {
		b[pos] &= ^(uint8(1) << j)
	}
}

func (b BitUint8) Len() int { return 8 * len(b) }

func (b BitUint8) WriteTo(w io.Writer) (n int64, err error) {
	sz := int64(b.Len())
	err = binary.Write(w, binary.LittleEndian, sz)
	if err != nil {
		return n, fmt.Errorf("cannot encode size of bitset: %w", err)
	}
	n += 8
	nw, err := w.Write(b)
	n += int64(nw)
	return n, err
}

func (b *BitUint8) ReadFrom(r io.Reader) (n int64, err error) {
	var sz int64
	err = binary.Read(r, binary.LittleEndian, &sz)
	if err != nil {
		return n, fmt.Errorf("cannot decode size of bitset: %w", err)
	}
	n += 8
	*b = NewBitUint8(int(sz))
	nr, err := io.ReadFull(r, *b)
	n += int64(nr)
	return n, err
}

func NewBitUint8(n int) BitUint8 {
	assert(n > 0, "n must be positive")

	return make(BitUint8, (n+7)/8)
}

type BitBool []bool

func (b BitBool) Has(i int) bool {
	if i >= len(b) {
		return false
	}
	return b[i]
}

func (b *BitBool) Set(i int, t bool) {
	if i >= len(*b) {
		b.grow(1 + i)
	}
	(*b)[i] = t
}

func (b *BitBool) grow(size int) {
	b2 := make(BitBool, size)
	copy(b2, *b)
	*b = b2
}

func (b BitBool) Len() int { return len(b) }

func assert(exp bool, format string) {
	if !exp {
		panic(format)
	}
}
