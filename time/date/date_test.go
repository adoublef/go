// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package date_test

import (
	"encoding/json"
	"strings"
	"testing"

	. "go.adoublef.dev/time/date"
)

func TestDate(t *testing.T) {
	t.Parallel()
	t.Run("JSON", func(t *testing.T) {
		t.Parallel()
		body := `"2016-10-08"`
		var d Date
		err := json.Unmarshal([]byte(body), &d)
		if err != nil {
			t.Errorf("Date.UnmarshalText: %v", err)
		}
		if got, want := d.Day, 8; got != want {
			t.Errorf("Date.Day: got=%d; want=%d", got, want)
		}
		if got, want := d.Month, October; got != want {
			t.Errorf("Date.Month: got=%q; want=%q", got, want)
		}
		if got, want := d.Year, 2016; got != want {
			t.Errorf("Date.Year: got=%d; want=%d", got, want)
		}
		out, _ := json.Marshal(d)
		if !strings.EqualFold(string(out), body) {
			t.Errorf("Date.MarshalText: got=%s; want=%s", string(out), body)
		}
	})
}
