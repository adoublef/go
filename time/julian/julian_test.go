// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package julian_test

import (
	"encoding/json"
	"strings"
	"testing"

	. "go.adoublef.dev/time/julian"
)

func TestTime(t *testing.T) {
	t.Parallel()
	t.Run("JSON", func(t *testing.T) {
		t.Parallel()
		body := `"2016-10-08"`
		var d Time
		err := json.Unmarshal([]byte(body), &d)
		if err != nil {
			t.Errorf("Time.UnmarshalText: %v", err)
		}
		// d.Time() check something
		out, _ := json.Marshal(d)
		if got, want := string(out), body[:len(body)-1]; !strings.HasPrefix(got, want) {
			t.Errorf("Time.MarshalText: got=%s; want=%s", string(out), body)
		}
	})
}
