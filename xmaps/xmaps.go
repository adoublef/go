package xmaps

import (
	"slices"

	"go.adoublef.dev/xiter"
)

func GroupFunc[K comparable, Slice ~[]V, V any](aa Slice, f func(V) K) map[K]Slice {
	return xiter.Reduce(func(m map[K]Slice, a V) map[K]Slice {
		k := f(a)
		m[k] = append(m[k], a)
		return m
	}, make(map[K]Slice), slices.Values(aa))
}

func SetFunc[K comparable, Slice ~[]V, V any](aa Slice, f func(V) K) map[K]struct{} {
	return xiter.Reduce(func(m map[K]struct{}, a V) map[K]struct{} {
		k := f(a)
		m[k] = struct{}{}
		return m
	}, make(map[K]struct{}), slices.Values(aa))
}
