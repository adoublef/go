// Copyright 2024 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xiota_test

import (
	"fmt"
	"testing"

	"go.adoublef.dev/xiota"
	. "go.adoublef.dev/xiota"
)

func ExampleFormat() {
	type A uint
	fmt.Println(Format[A](2, []string(nil), 0, 0, 0))
	// Output:
	// 	%!A(2)
}

type Weekday uint8

const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

var weekdayNames = []string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

type Color uint8

const (
	Red Color = iota + 1
	Green
	Blue
	Yellow
	Purple
)

var colorNames = []string{
	"Red",
	"Green",
	"Blue",
	"Yellow",
	"Purple",
}

type Direction uint8

const (
	North Direction = iota
	East
	South
	West
)

var directionNames = []string{
	"North",
	"East",
	"South",
	"West",
}

type Month uint8

const (
	January Month = iota + 1
	February
	March
	April
	May
)

var monthNames = []string{
	"January",
	"February",
	"March",
	"April",
	"May",
}

func TestFormat(t *testing.T) {
	type testcase struct {
		name     string
		value    any
		names    []string
		min      any
		max      any
		offset   any
		expected string
	}

	tt := []testcase{
		{
			name:     "Valid Weekday",
			value:    Monday,
			names:    weekdayNames,
			min:      Sunday,
			max:      Saturday,
			offset:   Sunday,
			expected: "Monday",
		},
		{
			name:     "Invalid Weekday",
			value:    Weekday(10),
			names:    weekdayNames,
			min:      Sunday,
			max:      Saturday,
			offset:   Sunday,
			expected: "%!Weekday(10)",
		},
		{
			name:     "Valid Color with offset=1",
			value:    Green,
			names:    colorNames,
			min:      Red,
			max:      Purple,
			offset:   Red, // With 1-based indexing, offset is min-1
			expected: "Green",
		},
		{
			name:     "Zero value with zero offset",
			value:    Sunday,
			names:    weekdayNames,
			min:      Sunday,
			max:      Saturday,
			offset:   Sunday, // Same as 0
			expected: "Sunday",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			switch value := tc.value.(type) {
			case Weekday:
				min := tc.min.(Weekday)
				max := tc.max.(Weekday)
				offset := tc.offset.(Weekday)
				result := Format(value, tc.names, min, max, offset)
				if result != tc.expected {
					t.Errorf("Format() = %v, want %v", result, tc.expected)
				}
			case Color:
				min := tc.min.(Color)
				max := tc.max.(Color)
				offset := tc.offset.(Color)
				result := Format(value, tc.names, min, max, offset)
				if result != tc.expected {
					t.Errorf("Format() = %v, want %v", result, tc.expected)
				}
			default:
				t.Errorf("Unsupported type in test case: %T", tc.value)
			}
		})
	}
}

func TestParse(t *testing.T) {
	type testcase struct {
		name          string
		input         string
		names         []string
		offset        int
		expectedType  any
		expectedValue any
		wantErr       bool
	}

	tests := []testcase{
		{
			name:          "Valid Direction",
			input:         "North",
			names:         directionNames,
			offset:        0,
			expectedType:  Direction(0),
			expectedValue: North,
			wantErr:       false,
		},
		{
			name:          "Valid Direction case insensitive",
			input:         "north",
			names:         directionNames,
			offset:        0,
			expectedType:  Direction(0),
			expectedValue: North,
			wantErr:       false,
		},
		{
			name:          "Invalid Direction",
			input:         "Northeast",
			names:         directionNames,
			offset:        0,
			expectedType:  Direction(0),
			expectedValue: Direction(0),
			wantErr:       false,
		},
		{
			name:          "Valid Month with offset",
			input:         "February",
			names:         monthNames,
			offset:        1, // Months start at 1
			expectedType:  Month(0),
			expectedValue: February,
			wantErr:       false,
		},
		{
			name:          "Empty string",
			input:         "",
			names:         directionNames,
			offset:        0,
			expectedType:  Direction(0),
			expectedValue: Direction(0),
			wantErr:       true,
		},
		{
			name:          "Partial match prefix",
			input:         "Northern",
			names:         directionNames,
			offset:        0,
			expectedType:  Direction(0),
			expectedValue: North,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.expectedType.(type) {
			case Direction:
				result, err := xiota.Parse[Direction](tt.names, tt.input, tt.offset)

				if (err != nil) != tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if !tt.wantErr && result != tt.expectedValue.(Direction) {
					t.Errorf("Parse() = %v, want %v", result, tt.expectedValue)
				}

			case Month:
				result, err := xiota.Parse[Month](tt.names, tt.input, tt.offset)

				if (err != nil) != tt.wantErr {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if !tt.wantErr && result != tt.expectedValue.(Month) {
					t.Errorf("Parse() = %v, want %v", result, tt.expectedValue)
				}

			default:
				t.Fatalf("Unsupported type in test case: %T", tt.expectedType)
			}
		})
	}
}
