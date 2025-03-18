// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package consistenthash provides an implementation of consistent hashing.
package consistenthash

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Hasher defines an interface for hash functions that produce uint32 values.
type Hasher interface {
	Hash(b []byte) uint32
}

// HashFunc is a function type that implements the Hasher interface.
type HashFunc func(b []byte) uint32

// Hash implements the Hasher interface for HashFunc.
func (hf HashFunc) Hash(b []byte) uint32 {
	return hf(b)
}

// Map implements consistent hashing ring.
type Map struct {
	Hasher Hasher
	k      int   // k number of sets
	keys   []int // sorted
	set    map[int]string
}

// Add adds some keys to the hash.
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := range m.k {
			h := int(m.Hasher.Hash([]byte((itoa(i) + key))))
			j := len(m.keys)
			m.keys = append(m.keys, h)
			for j > 0 && m.keys[j-1] > h {
				m.keys[j] = m.keys[j-1]
				j--
			}
			m.keys[j] = h
			m.set[h] = key
		}
	}
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if m.IsEmpty() {
		return ""
	}

	h := int(m.Hasher.Hash([]byte(key)))
	idx, _ := search(m.keys, h)
	if idx == len(m.keys) {
		idx = 0
	}
	return m.set[m.keys[idx]]
}

// Returns true if there are no items available.
func (m *Map) IsEmpty() bool {
	return len(m.keys) == 0
}

// WriteTo implements io.WriterTo.
func (m Map) WriteTo(w io.Writer) (n int64, err error) {
	k := int64(m.k)
	err = binary.Write(w, binary.LittleEndian, k)
	if err != nil {
		return n, fmt.Errorf("cannot encode size of replica: %w", err)
	}
	n += 8

	nk := int64(len(m.keys))
	err = binary.Write(w, binary.LittleEndian, nk)
	if err != nil {
		return n, fmt.Errorf("cannot encode size of keys: %w", err)
	}
	n += 8
	for _, k := range m.keys {
		err = binary.Write(w, binary.LittleEndian, int64(k))
		if err != nil {
			return n, fmt.Errorf("cannot encode key %d: %w", k, err)
		}
		n += 8
	}

	ns := int64(len(m.set))
	err = binary.Write(w, binary.LittleEndian, ns)
	if err != nil {
		return n, fmt.Errorf("cannot encode size of set: %w", err)
	}
	n += 8
	for k, v := range m.set {
		err = binary.Write(w, binary.LittleEndian, int64(k))
		if err != nil {
			return n, fmt.Errorf("cannot encode key %d: %w", k, err)
		}
		n += 8

		err = binary.Write(w, binary.LittleEndian, int64(len(v)))
		if err != nil {
			return n, err
		}
		n += 8
		nw, err := w.Write([]byte(v))
		if err != nil {
			return n, err
		}
		n += int64(nw)
	}
	return n, nil
}

// ReadFrom implements io.ReaderFrom.
func (m *Map) ReadFrom(r io.Reader) (n int64, err error) {
	var k int64
	err = binary.Read(r, binary.LittleEndian, &k)
	if err != nil {
		return 0, fmt.Errorf("cannot read size of replica: %w", err)
	}
	n += 8
	m.k = int(k)

	var nk int64
	err = binary.Read(r, binary.LittleEndian, &nk)
	if err != nil {
		return n, fmt.Errorf("cannot read size of keys: %w", err)
	}
	n += 8
	m.keys = make([]int, nk)
	for i := range m.keys {
		var key int64
		err := binary.Read(r, binary.LittleEndian, &key)
		if err != nil {
			return n, fmt.Errorf("cannot read key: %w", err)
		}
		m.keys[i] = int(key)
		n += 8
	}

	var ns int64
	err = binary.Read(r, binary.LittleEndian, &ns)
	if err != nil {
		return n, fmt.Errorf("cannot read size of set: %w", err)
	}
	n += 8
	m.set = make(map[int]string, ns)
	for i := int64(0); i < ns; i++ {
		var key int64
		err := binary.Read(r, binary.LittleEndian, &key)
		if err != nil {
			return n, fmt.Errorf("cannot read key: %w", err)
		}
		n += 8

		var ns int64
		err = binary.Read(r, binary.LittleEndian, &ns)
		if err != nil {
			return n, fmt.Errorf("cannot read size of value: %w", err)
		}
		n += 8
		buf := make([]byte, ns)
		read, err := io.ReadFull(r, buf)
		if err != nil {
			return n, err
		}
		n += int64(read)

		m.set[int(key)] = string(buf)
	}

	return n, nil
}

// New creates a new [Map].
func New(k int, h Hasher) *Map {
	assert(k > 0, "k must be greater than zero")
	assert(h != nil, "hasher cannot be nil")

	return &Map{k: k, Hasher: h, set: make(map[int]string)}
}

func itoa(n int) string {
	var buf [20]byte // max uint64 length
	i := len(buf)
	// Process two digits at a time
	for n >= 100 {
		q := n / 100
		r := n - q*100 // remainder

		d1 := r / 10    // first digit
		d2 := r - d1*10 // second digit

		i -= 2
		buf[i] = '0' + byte(d1)
		buf[i+1] = '0' + byte(d2)
		n = q
	}
	// Handle last 1-2 digits
	if n >= 10 {
		d1 := n / 10
		d2 := n - d1*10
		i -= 2
		buf[i] = '0' + byte(d1)
		buf[i+1] = '0' + byte(d2)
	} else {
		i--
		buf[i] = '0' + byte(n)
	}
	return string(buf[i:])
}

func search(x []int, target int) (int, bool) {
	l, r := 0, len(x)
	// Binary search invariant:
	// - All elements to the left of 'left' are < target
	// - All elements starting from 'right' are >= target
	for l < r {
		// Calculate midpoint safely without overflow
		mid := l + (r-l)/2

		if x[mid] < target {
			l = mid + 1
		} else {
			r = mid
		}
	}

	// Check if target was found
	if l < len(x) && x[l] == target {
		return l, true
	}
	return l, false
}

func assert(exp bool, format string) {
	if !exp {
		panic(format)
	}
}
