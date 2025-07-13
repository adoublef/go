// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xprof_test

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "go.adoublef.dev/runtime/xprof"
	"go.adoublef.dev/testing/is"
)

type checkFn func(t *testing.T, stdout, stderr []byte, err error)

func TestProfile(t *testing.T) {
	f, err := os.CreateTemp("", "profile_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	var profileTests = []struct {
		name   string
		code   string
		checks []checkFn
	}{
		{
			name: "default profile (cpu)",
			code: `
package main

import "go.adoublef.dev/runtime/xprof"

func main() {
	defer xprof.Start("").Stop()
}	
`,
			checks: []checkFn{
				NoStdout,
				Stderr("xprof: cpu profiling enabled"),
				NoErr,
			},
		},
		{
			name: "memory profile",
			code: `
package main

import "go.adoublef.dev/runtime/xprof"

func main() {
	defer xprof.Start("", xprof.Mem).Stop()
}	
`,
			checks: []checkFn{
				NoStdout,
				Stderr("xprof: memory profiling enabled"),
				NoErr,
			},
		},
		{
			name: "memory profile (rate 2048)",
			code: `
package main

import "go.adoublef.dev/runtime/xprof"

func main() {
	defer xprof.Start("", xprof.MemRate(2048)).Stop()
}	
`,
			checks: []checkFn{
				NoStdout,
				Stderr("xprof: memory profiling enabled (rate 2048)"),
				NoErr,
			},
		},
		{
			name: "double start",
			code: `
package main

import "go.adoublef.dev/runtime/xprof"

func main() {
	xprof.Start("")
	xprof.Start("")
}	
`,
			checks: []checkFn{
				NoStdout,
				Stderr("cpu profiling enabled", "xprof: Start() already called"),
				Err,
			},
		},
		{
			name: "block profile",
			code: `
package main

import "go.adoublef.dev/runtime/xprof"

func main() {
	defer xprof.Start("", xprof.Block).Stop()
}	
`,
			checks: []checkFn{
				NoStdout,
				Stderr("xprof: block profiling enabled"),
				NoErr,
			},
		},
		{
			name: "mutex profile",
			code: `
package main

import "go.adoublef.dev/runtime/xprof"

func main() {
	defer xprof.Start("", xprof.Mutex).Stop()
}
`,
			checks: []checkFn{
				NoStdout,
				Stderr("xprof: mutex profiling enabled"),
				NoErr,
			},
		},
		{
			name: "profile path error",
			code: `
package main

import "go.adoublef.dev/runtime/xprof"

func main() {
		defer xprof.Start("` + f.Name() + `").Stop()
}	
`,
			checks: []checkFn{
				NoStdout,
				Stderr("could not create initial output"),
				Err,
			},
		},
		{
			name: "multiple profile sessions",
			code: `
package main

import "go.adoublef.dev/runtime/xprof"

func main() {
	xprof.Start("", xprof.CPU).Stop()
	xprof.Start("", xprof.Mem).Stop()
	xprof.Start("", xprof.Block).Stop()
	xprof.Start("", xprof.CPU).Stop()
	xprof.Start("", xprof.Mutex).Stop()
}
`,
			checks: []checkFn{
				NoStdout,
				Stderr("xprof: cpu profiling enabled",
					"xprof: cpu profiling disabled",
					"xprof: memory profiling enabled",
					"xprof: memory profiling disabled",
					"xprof: block profiling enabled",
					"xprof: block profiling disabled",
					"xprof: cpu profiling enabled",
					"xprof: cpu profiling disabled",
					"xprof: mutex profiling enabled",
					"xprof: mutex profiling disabled"),
				NoErr,
			},
		},
		{
			name: "profile quiet",
			code: `
package main

import "go.adoublef.dev/runtime/xprof"

func main() {
        defer xprof.Start("",xprof.Quiet).Stop()
}       
`,
			checks: []checkFn{NoStdout, NoStderr, NoErr},
		},
	}
	for _, tt := range profileTests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := runTest(t, tt.code)
			for _, f := range tt.checks {
				f(t, stdout, stderr, err)
			}
		})
	}
}

// NoStdout checks that stdout was blank.
func NoStdout(t *testing.T, stdout, _ []byte, _ error) {
	t.Helper()
	is.Equal(t, 0, len(stdout)) // stdout: wanted 0 bytes
}

// Stderr verifies that the given lines match the output from stderr
func Stderr(lines ...string) checkFn {
	return func(t *testing.T, _, stderr []byte, _ error) {
		r := bytes.NewReader(stderr)
		if !validateOutput(r, lines) {
			t.Errorf("stderr: wanted '%s', got '%s'", lines, stderr)
		}
	}
}

// NoStderr checks that stderr was blank.
func NoStderr(t *testing.T, _, stderr []byte, _ error) {
	t.Helper()
	is.Equal(t, 0, len(stderr)) // stderr: wanted 0 bytes
}

// Err checks that there was an error returned
func Err(t *testing.T, _, _ []byte, err error) {
	t.Helper()
	is.True(t, err != nil) // expected error
}

// NoErr checks that err was nil
func NoErr(t *testing.T, _, _ []byte, err error) {
	t.Helper()
	is.OK(t, err) // unexpected error
}

// validatedOutput validates the given slice of lines against data from the given reader.
func validateOutput(r io.Reader, want []string) bool {
	s := bufio.NewScanner(r)
	for _, line := range want {
		if !s.Scan() || !strings.Contains(s.Text(), line) {
			return false
		}
	}
	return true
}

var validateOutputTests = []struct {
	input string
	lines []string
	want  bool
}{{
	input: "",
	want:  true,
}, {
	input: `xprof: yes
`,
	want: true,
}, {
	input: `xprof: yes
`,
	lines: []string{"xprof: yes"},
	want:  true,
}, {
	input: `xprof: yes
xprof: no
`,
	lines: []string{"xprof: yes"},
	want:  true,
}, {
	input: `xprof: yes
xprof: no
`,
	lines: []string{"xprof: yes", "xprof: no"},
	want:  true,
}, {
	input: `xprof: yes
xprof: no
`,
	lines: []string{"xprof: no"},
	want:  false,
}}

func TestValidateOutput(t *testing.T) {
	for _, tt := range validateOutputTests {
		r := strings.NewReader(tt.input)
		got := validateOutput(r, tt.lines)
		if tt.want != got {
			t.Errorf("validateOutput(%q, %q), want %v, got %v", tt.input, tt.lines, tt.want, got)
		}
	}
}

// runTest executes the go program supplied and returns the contents of stdout,
// stderr, and an error which may contain status information about the result
// of the program.
func runTest(t *testing.T, code string) ([]byte, []byte, error) {
	t.Helper()

	gopath, err := os.MkdirTemp("", "profile-gopath")
	is.OK(t, err) // MkdirTemp
	defer os.RemoveAll(gopath)

	srcdir := filepath.Join(gopath, "src")
	err = os.Mkdir(srcdir, 0755)
	is.OK(t, err) // Mkdir

	src := filepath.Join(srcdir, "main.go")
	err = os.WriteFile(src, []byte(code), 0644)
	is.OK(t, err) // WriteFile

	cmd := exec.Command("go", "run", src)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func ExampleStart() {
	// start a simple CPU profile and register
	// a defer to Stop (flush) the profiling data.
	defer Start("").Stop()
}

func ExampleCPU() {
	// CPU profiling is the default profiling mode, but you can specify it
	// explicitly for completeness.
	defer Start("", CPU).Stop()
}

func ExampleMem() {
	// use memory profiling, rather than the default cpu profiling.
	defer Start("", Mem).Stop()
}

func ExampleMemRate() {
	// use memory profiling with custom rate.
	defer Start("", MemRate(2048)).Stop()
}

func ExampleMemHeap() {
	// use heap memory profiling.
	defer Start("", MemHeap).Stop()
}

func ExampleMemAllocs() {
	// use allocs memory profiling.
	defer Start("", MemAllocs).Stop()
}

func ExampleNoShutdownHook() {
	// disable the automatic shutdown hook.
	defer Start("", NoShutdownHook).Stop()
}

func ExampleStart_withFlags() {
	// use the flags package to selectively enable profiling.
	mode := flag.String("mode", "", "enable profiling mode, one of [cpu, mem, mutex, block]")
	flag.Parse()
	switch *mode {
	case "cpu":
		defer Start("", CPU).Stop()
	case "mem":
		defer Start("", Mem).Stop()
	case "mutex":
		defer Start("", Mutex).Stop()
	case "block":
		defer Start("", Block).Stop()
	default:
		// do nothing
	}
}

func ExampleTrace() {
	// use execution tracing, rather than the default cpu profiling.
	defer Start("", Trace).Stop()
}
