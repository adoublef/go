// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package du provides disk usage measurement representation.
// It defines a Size type (uint64) that can be converted to and from
// human-readable strings with units (K, M, G, etc.).
package du

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
)

type Size uint64

const (
	B Size = 1 << (10 * iota)
	K
	M
	G
	T
	P
	E
)

func (s Size) Int() int { return int(s) }

func (s Size) String() string {
	if s < K {
		return fmt.Sprintf("%dB", s)
	}
	div, exp := K, 0
	for n := s / K; n >= K; n /= K {
		div *= K
		exp++
	}
	// NOTE look to using strings.Builder instead
	return fmt.Sprintf("%.2f%c", float64(s)/float64(div), "KMGTPE"[exp])
}

// Set implements the flag.Value interface
func (s *Size) Set(value string) (err error) {
	*s, err = ParseSize(value)
	if err != nil {
		return err
	}
	return nil
}

// Scan implements the sql.Scanner interface for database deserialization
func (s *Size) Scan(src any) (err error) {
	switch v := src.(type) {
	case int64: // can other values be used?
		*s = Size(v)
		return nil
	case []byte, string:
		// Parse string representations
		var str string
		if strBytes, ok := v.([]byte); ok {
			str = string(strBytes)
		} else {
			str = v.(string)
		}
		if val, err := strconv.ParseUint(str, 10, 64); err == nil {
			*s = Size(val)
			return nil
		}
		*s, err = ParseSize(str)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported type %T for Size", src)
	}
}

// Value implements the driver.Valuer interface for database serialization
func (s Size) Value() (driver.Value, error) {
	return int64(s), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (s *Size) UnmarshalText(text []byte) (err error) {
	if val, err := strconv.ParseUint(string(text), 10, 64); err == nil {
		*s = Size(val)
		return nil
	}
	*s, err = ParseSize(string(text))
	if err != nil {
		return err
	}
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface
func (s Size) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatUint(uint64(s), 10)), nil
}

func ParseSize(s string) (Size, error) {
	if len(s) == 0 {
		return 0, nil
	}

	var i int
	for i = 0; i < len(s) && ((s[i] >= '0' && s[i] <= '9') || s[i] == '.'); i++ {
	}
	num, unit := s[:i], s[i:]

	if len(unit) > 2 {
		return 0, fmt.Errorf("invalid unit: %q", unit)
	}
	n, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format: %s", err)
	}
	var multi Size
	switch unit {
	case "", "b", "B": // Bytes
		multi = B
	case "k", "K", "kb", "KB":
		multi = K
	case "m", "M", "mb", "MB":
		multi = M
	case "g", "G", "gb", "GB":
		multi = G
	case "t", "T", "tb", "TB":
		multi = T
	case "p", "P", "pb", "PB":
		multi = P
	case "e", "E", "eb", "EB":
		multi = E
	default:
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}
	size := Size(n * float64(multi))
	return size, nil
}

var (
	ErrSizeSyntax = errors.New("bad size syntax")
)
