// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"embed"
	"net/http"
)

//go:embed all:*.html
var fsys embed.FS

func main() {
	http.Handle("GET /{$}", http.FileServerFS(fsys))
	http.ListenAndServe(":8080", nil)
}
