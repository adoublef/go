// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package du_test

import (
	"fmt"
	"strings"
	"testing"

	. "go.adoublef.dev/os/du"
)

func TestSize(t *testing.T) {
	type testcase struct {
		input   string
		wantErr bool
	}

	tt := []testcase{
		{"", false},
		{"1", false},
		{"1g", false},
		{"1G", false},
		{"1gb", false},
		{"1GB", false},
		{"1.0GB", false},
		{"1 GB", true},
	}

	// Test UnmarshalText
	t.Run("UnmarshalText", func(t *testing.T) {
		for _, tc := range tt {
			t.Run(tc.input, func(t *testing.T) {
				var s Size
				err := s.UnmarshalText([]byte(tc.input))
				if (err != nil) != tc.wantErr {
					t.Errorf("UnmarshalText unexpected error: got %v, wantErr %v", err, tc.wantErr)
				}
			})
		}
	})

	// Test Set
	t.Run("Set", func(t *testing.T) {
		for _, tc := range tt {
			t.Run(tc.input, func(t *testing.T) {
				var s Size
				err := s.Set(tc.input)
				if (err != nil) != tc.wantErr {
					t.Errorf("Set unexpected error: got %v, wantErr %v", err, tc.wantErr)
				}
			})
		}
	})

	// Test Scan with string
	t.Run("ScanString", func(t *testing.T) {
		for _, tc := range tt {
			t.Run(tc.input, func(t *testing.T) {
				var s Size
				err := s.Scan(tc.input)
				if (err != nil) != tc.wantErr {
					t.Errorf("Scan(string) unexpected error: got %v, wantErr %v", err, tc.wantErr)
				}
			})
		}
	})

	// Test Scan with []byte
	t.Run("ScanBytes", func(t *testing.T) {
		for _, tc := range tt {
			t.Run(tc.input, func(t *testing.T) {
				var s Size
				err := s.Scan([]byte(tc.input))
				if (err != nil) != tc.wantErr {
					t.Errorf("Scan([]byte) unexpected error: got %v, wantErr %v", err, tc.wantErr)
				}
			})
		}
	})

	// Test MarshalText
	t.Run("MarshalText", func(t *testing.T) {
		sizes := []uint64{0, 1024, 1048576, 1073741824}
		for _, size := range sizes {
			t.Run(fmt.Sprintf("%d", size), func(t *testing.T) {
				s := Size(size)
				data, err := s.MarshalText()
				if err != nil {
					t.Errorf("MarshalText error: %v", err)
				}

				// Unmarshal back and check
				var s2 Size
				err = s2.UnmarshalText(data)
				if err != nil {
					t.Errorf("UnmarshalText error: %v", err)
				}

				if s != s2 {
					t.Errorf("Marshal/Unmarshal roundtrip failed: got %d, want %d", s2, s)
				}
			})
		}
	})

	// Test Scan with numeric types
	t.Run("ScanNumericTypes", func(t *testing.T) {
		// Only testing int64 as per error message
		val := int64(300)
		t.Run("int64", func(t *testing.T) {
			var s Size
			err := s.Scan(val)
			if err != nil {
				t.Errorf("Scan error: %v", err)
				return
			}

			if uint64(s) != uint64(val) {
				t.Errorf("Scan incorrect value: got %d, want %d", s, val)
			}
		})

		// Test with unsupported numeric types
		unsupportedTypes := []interface{}{
			int(100),
			int32(200),
			uint(400),
			uint32(500),
			uint64(600),
		}

		for _, val := range unsupportedTypes {
			t.Run(fmt.Sprintf("%T", val), func(t *testing.T) {
				var s Size
				err := s.Scan(val)
				if err == nil {
					t.Errorf("Expected error for unsupported type %T, got nil", val)
				}

				if !strings.Contains(err.Error(), "unsupported type") {
					t.Errorf("Expected 'unsupported type' error, got: %v", err)
				}
			})
		}
	})

	// Test Value method
	t.Run("Value", func(t *testing.T) {
		s := Size(12345)
		val, err := s.Value()
		if err != nil {
			t.Errorf("Value() error: %v", err)
		}

		// Value should return an int64
		i64, ok := val.(int64)
		if !ok {
			t.Errorf("Value() returned %T, want int64", val)
		}

		if i64 != 12345 {
			t.Errorf("Value() returned %d, want %d", i64, 12345)
		}
	})

	// Test unsupported type error
	t.Run("UnsupportedType", func(t *testing.T) {
		var s Size
		err := s.Scan(true) // Boolean is unsupported
		if err == nil {
			t.Errorf("Expected error for unsupported type, got nil")
		}

		if !strings.Contains(err.Error(), "unsupported type") {
			t.Errorf("Expected 'unsupported type' error, got: %v", err)
		}
	})
}

func TestParseSize(t *testing.T) {
	type testcase struct {
		s       string
		wantErr bool
	}

	tt := []testcase{
		{"", false},
		{"1", false},
		{"1g", false},
		{"1G", false},
		{"1gb", false},
		{"1GB", false},
		{"1.0GB", false},
		{"1 GB", true},
	}

	for _, tc := range tt {
		t.Run(tc.s, func(t *testing.T) {
			_, err := ParseSize(tc.s)
			if (err == nil) == tc.wantErr {
				t.Errorf("unexpected error: got %v", err)
			}
		})
	}
}
