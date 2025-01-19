// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cmp"
	"embed"
	"net/http"
	"os"
)

//go:embed all:*.html
var fsys embed.FS

func main() {
	http.Handle("GET /{$}", http.FileServerFS(fsys))
	addr := ":" + cmp.Or(os.Getenv("PORT"), "8080")
	http.ListenAndServe(addr, nil)
}
