// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package du
package du

import (
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
	default:
		return 0, fmt.Errorf("unknown size unit: %s", unit)
	}
	size := Size(n * float64(multi))
	return size, nil
}

var (
	ErrSizeSyntax = errors.New("bad size syntax")
)
