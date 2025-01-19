// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package consistenthash_test

import (
	"hash/crc32"
	"io"
	"math"
	"net"
	"strconv"
	"testing"

	. "go.adoublef.dev/container/consistenthash"
	"golang.org/x/exp/rand"
)

func TestMap(t *testing.T) {
	t.Parallel()
	t.Run("Hashing", func(t *testing.T) {
		t.Parallel()

		// Override the hash function to return easier to reason about values. Assumes
		// the keys can be converted to an integer.
		hash := New(3, HashFunc(func(key []byte) uint32 {
			i, err := strconv.Atoi(string(key))
			if err != nil {
				panic(err)
			}
			return uint32(i)
		}))

		// Given the above hash function, this will give replicas with "hashes":
		// 2, 4, 6, 12, 14, 16, 22, 24, 26
		hash.Add("6", "4", "2")

		testCases := map[string]string{
			"2":  "2",
			"11": "2",
			"23": "4",
			"27": "2",
		}

		for k, v := range testCases {
			if got := hash.Get(k); got != v {
				t.Errorf("Asking for %s, should have yielded %s; got %s instead", k, v, got)
			}
		}

		// Adds 8, 18, 28
		hash.Add("8")

		// 27 should now map to 8.
		testCases["27"] = "8"

		for k, v := range testCases {
			if got := hash.Get(k); got != v {
				t.Errorf("Asking for %s, should have yielded %s; got %s instead", k, v, got)
			}
		}
	})
	t.Run("Consistency", func(t *testing.T) {
		t.Parallel()

		hash1 := New(1, HashFunc(crc32.ChecksumIEEE))
		hash2 := New(1, HashFunc(crc32.ChecksumIEEE))

		hash1.Add("Bill", "Bob", "Bonny")
		hash2.Add("Bob", "Bonny", "Bill")

		if hash1.Get("Ben") != hash2.Get("Ben") {
			t.Errorf("Fetching 'Ben' from both hashes should be the same")
		}

		hash2.Add("Becky", "Ben", "Bobby")

		if hash1.Get("Ben") != hash2.Get("Ben") ||
			hash1.Get("Bob") != hash2.Get("Bob") ||
			hash1.Get("Bonny") != hash2.Get("Bonny") {
			t.Errorf("Direct matches should always return the same entry")
		}
	})
	t.Run("Distrubtion", func(t *testing.T) {
		t.Parallel()

		hosts := []string{"a.svc.local", "b.svc.local", "c.svc.local"}
		const cases = 10000

		strings := make([]string, cases)

		for i := 0; i < cases; i++ {
			r := rand.Int31()
			ip := net.IPv4(192, byte(r>>16), byte(r>>8), byte(r))
			strings[i] = ip.String()
		}

		hashFuncs := map[string]Hasher{
			"crc32": HashFunc(crc32.ChecksumIEEE),
		}

		for name, hashFunc := range hashFuncs {
			t.Run(name, func(t *testing.T) {
				hash := New(512, hashFunc)
				hostMap := map[string]int{}

				for _, host := range hosts {
					hash.Add(host)
					hostMap[host] = 0
				}

				for i := range strings {
					host := hash.Get(strings[i])
					hostMap[host]++
				}

				// Calculate mean
				var sum float64
				percentages := make([]float64, len(hosts))
				for i, host := range hosts {
					percent := float64(hostMap[host]) / cases
					percentages[i] = percent
					sum += percent
				}
				mean := sum / float64(len(hosts))

				// Calculate standard deviation
				var sumSquares float64
				for _, percent := range percentages {
					diff := percent - mean
					sumSquares += diff * diff
				}
				sd := math.Sqrt(sumSquares / float64(len(hosts)))

				t.Logf("Standard Deviation: %.4f", sd)

				for host, a := range hostMap {
					t.Logf("host: %s, percent: %f", host, float64(a)/cases)
				}
			})
		}
	})

	t.Run("ReadFrom", func(t *testing.T) {
		t.Parallel()

		a := New(1, HashFunc(crc32.ChecksumIEEE))

		a.Add("Bill", "Bob", "Bonny")

		pr, pw := io.Pipe()
		defer pr.Close()
		go func() {
			defer pw.Close()
			_, err := a.WriteTo(pw)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
		}()

		var b Map
		b.Hasher = HashFunc(crc32.ChecksumIEEE)
		_, err := b.ReadFrom(pr)
		if err != nil {
			t.Errorf("Map.ReadFrom: %v", err)
		}

		if a.Get("Ben") != b.Get("Ben") ||
			a.Get("Bob") != b.Get("Bob") ||
			a.Get("Bonny") != b.Get("Bonny") {
			t.Errorf("Direct matches should always return the same entry")
		}
	})
}
