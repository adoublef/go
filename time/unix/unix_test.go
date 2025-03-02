// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix_test

import (
	"testing"

	. "go.adoublef.dev/time/unix"
)

func TestTime(t *testing.T) {
	t.Parallel()
	t.Run("SQL", func(t *testing.T) {
		t.Parallel()
		body := `2016-10-08 16:04:05`
		var d Time
		// Test Scan
		err := d.Scan(body)
		if err != nil {
			t.Errorf("Time.Scan: %v", err)
		}

		// Test Value
		got, err := d.Value()
		if err != nil {
			t.Errorf("Time.Value: %v", err)
		}

		if got, want := got.(int64), int64(1475942645); got != want {
			t.Errorf("Time.Value: got=%d; want=%d", got, want)
		}
	})
}
