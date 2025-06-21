// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xhttp

import "context"

type contextKey struct{ s string }

func (k contextKey) String() string { return "go.adoublef.dev/net/xhttp: " + k.s }

// mustValue returns the context value else panics.
func mustValue[E any](ctx context.Context, key any) E {
	v, ok := value[E](ctx, key)
	if !ok {
		panic("context is missing")
	}
	return v
}

// value returns the context value.
func value[E any](ctx context.Context, key any) (E, bool) {
	v, ok := ctx.Value(key).(E)
	return v, ok
}
