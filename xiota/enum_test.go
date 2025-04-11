// Copyright 2024 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xiota_test

import (
	"fmt"

	. "go.adoublef.dev/xiota"
)

func ExampleFormat() {
	type A uint
	fmt.Println(Format[A](nil, 2))
	// Output:
	// 	%!A(2)
}
