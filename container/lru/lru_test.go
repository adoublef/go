// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lru_test

import (
	"fmt"
	"testing"

	. "go.adoublef.dev/container/lru"
)

var getTests = []struct {
	name       string
	keyToAdd   string
	keyToGet   string
	expectedOk bool
}{
	{"string_hit", "myKey", "myKey", true},
	{"string_miss", "myKey", "nonsense", false},
}

func TestLRU(t *testing.T) {
	t.Parallel()
	t.Run("Get", func(t *testing.T) {
		t.Parallel()
		for _, tt := range getTests {
			lru := New[string, int](0)
			lru.Add(tt.keyToAdd, 1234)
			val, ok := lru.Get(tt.keyToGet)
			if ok != tt.expectedOk {
				t.Fatalf("%s: cache hit = %v; want %v", tt.name, ok, !ok)
			} else if ok && val != 1234 {
				t.Fatalf("%s expected get to return 1234 but got %v", tt.name, val)
			}
		}
	})
	t.Run("Remove", func(t *testing.T) {
		t.Parallel()
		lru := New[string, int](0)
		lru.Add("myKey", 1234)
		if val, ok := lru.Get("myKey"); !ok {
			t.Fatal("TestRemove returned no match")
		} else if val != 1234 {
			t.Fatalf("TestRemove failed.  Expected %d, got %v", 1234, val)
		}

		lru.Remove("myKey")
		if _, ok := lru.Get("myKey"); ok {
			t.Fatal("TestRemove returned a removed entry")
		}
	})

	t.Run("Evict", func(t *testing.T) {
		t.Parallel()
		evictedKeys := make([]string, 0)
		onEvictedFun := func(key string, value int) {
			evictedKeys = append(evictedKeys, key)
		}

		lru := New[string, int](20)
		lru.OnEvicted = onEvictedFun
		for i := 0; i < 22; i++ {
			lru.Add(fmt.Sprintf("myKey%d", i), 1234)
		}

		if len(evictedKeys) != 2 {
			t.Fatalf("got %d evicted keys; want 2", len(evictedKeys))
		}
		if evictedKeys[0] != "myKey0" {
			t.Fatalf("got %v in first evicted key; want %s", evictedKeys[0], "myKey0")
		}
		if evictedKeys[1] != "myKey1" {
			t.Fatalf("got %v in second evicted key; want %s", evictedKeys[1], "myKey1")
		}
	})
}
