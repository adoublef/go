// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !debug

package debug

// Printf calls Output to print to the standard logger.
func Printf(format string, v ...any) { /* no-op */ }
