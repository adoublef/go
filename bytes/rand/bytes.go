// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rand

import (
	"math/rand/v2"
	"unsafe"
)

const (
	set  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bits = 6           // 6 bits to represent a letter index
	mask = 1<<bits - 1 // All 1-bits, as many as letterIdxBits
	max  = 63 / bits   // # of letter indices fitting in 63 bits
)

// BytesN returns a random byte slice of length n
func BytesN(n int) []byte {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, rand.Int64(), max; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int64(), max
		}
		if idx := int(cache & mask); idx < len(set) {
			b[i] = set[idx]
			i--
		}
		cache >>= bits
		remain--
	}

	return b
}

// StringN returns a random string of length n
func StringN(n int) string {
	b := BytesN(n)
	return *(*string)(unsafe.Pointer(&b)) // equiv. string(b)
}
