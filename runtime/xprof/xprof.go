// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Copyright (c) 2013 Dave Cheney. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//   - Redistributions of source code must retain the above copyright
//
// notice, this list of conditions and the following disclaimer.
//   - Redistributions in binary form must reproduce the above
//
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package xprof.
//
// Extends the original source code https://pkg.go.dev/github.com/pkg/profile
package xprof

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync/atomic"
)

const (
	modeCpu = iota
	modeMem
	modeMutex
	modeBlock
	modeTrace
	modeThreadCreate
	modeGoroutine
)

type Profile struct {
	// quiet suppresses informational messages during profiling.
	quiet bool
	// noShutdownHook controls whether the profiling package should
	// hook SIGINT to write profiles cleanly.
	noShutdownHook bool
	// mode holds the type of profiling that will be made
	mode int
	// memProfileRate holds the rate for the memory profile.
	memProfileRate int
	// memProfileType holds the profile type for memory
	// profiles. Allowed values are `heap` and `allocs`.
	memProfileType string
	// closer holds a cleanup function that run after each profile
	closer func()
	// stopped records if a call to profile.Stop has been made
	stopped uint32
}

// NoShutdownHook controls whether the profiling package should
// hook SIGINT to write profiles cleanly.
// Programs with more sophisticated signal handling should set
// this to true and ensure the Stop() function returned from Start()
// is called during shutdown.
func NoShutdownHook(p *Profile) { p.noShutdownHook = true }

// Quiet suppresses informational messages during profiling.
func Quiet(p *Profile) { p.quiet = true }

// CPU enables cpu profiling.
// It disables any previous profiling settings.
func CPU(p *Profile) { p.mode = modeCpu }

// DefaultMemProfileRate is the default memory profiling rate.
// See also http://golang.org/pkg/runtime/#pkg-variables
const DefaultMemProfileRate = 4096

// Mem enables memory profiling.
// It disables any previous profiling settings.
func Mem(p *Profile) {
	p.memProfileRate = DefaultMemProfileRate
	p.mode = modeMem
}

// MemRate enables memory profiling at the preferred rate.
// It disables any previous profiling settings.
func MemRate(rate int) func(*Profile) {
	return func(p *Profile) {
		p.memProfileRate = rate
		p.mode = modeMem
	}
}

// MemHeap changes which type of memory profiling to profile
// the heap.
func MemHeap(p *Profile) {
	p.memProfileType = "heap"
	p.mode = modeMem
}

// MemAllocs changes which type of memory to profile
// allocations.
func MemAllocs(p *Profile) {
	p.memProfileType = "allocs"
	p.mode = modeMem
}

// Mutex enables mutex profiling.
// It disables any previous profiling settings.
func Mutex(p *Profile) { p.mode = modeMutex }

// Block enables block (contention) profiling.
// It disables any previous profiling settings.
func Block(p *Profile) { p.mode = modeBlock }

// Trace profile enables execution tracing.
// It disables any previous profiling settings.
func Trace(p *Profile) { p.mode = modeTrace }

// Thread enables thread creation profiling..
// It disables any previous profiling settings.
func Thread(p *Profile) { p.mode = modeThreadCreate }

// Go enables goroutine profiling.
// It disables any previous profiling settings.
func Go(p *Profile) { p.mode = modeGoroutine }

func (p *Profile) Stop() {
	if !atomic.CompareAndSwapUint32(&p.stopped, 0, 1) {
		return
	}
	p.closer()
	atomic.StoreUint32(&started, 0)
}

// started is non zero if a profile is running
var started uint32

// Start starts a new profiling session.
// The caller should call the Stop method on the value returned
// to cleanly stop profiling.
func Start(path string, options ...func(*Profile)) interface {
	Stop()
} {
	if !atomic.CompareAndSwapUint32(&started, 0, 1) {
		log.Fatal("xprof: Start() already called")
	}

	var prof Profile
	for _, o := range options {
		o(&prof)
	}

	path, err := func() (string, error) {
		if p := path; p != "" {
			return p, os.MkdirAll(p, 0777)
		}
		return os.MkdirTemp("", "profile")
	}()
	if err != nil {
		log.Fatalf("xprof: could not create initial output directory: %v", err)
	}
	logf := func(format string, args ...interface{}) {
		if !prof.quiet {
			log.Printf(format, args...)
		}
	}

	if prof.memProfileType == "" {
		prof.memProfileType = "heap"
	}
	switch prof.mode {
	case modeCpu:
		fn := filepath.Join(path, "cpu.pprof")
		f, err := os.Create(fn)
		if err != nil {
			log.Fatalf("xprof: could not create cpu profile %q: %v", fn, err)
		}
		logf("xprof: cpu profiling enabled, %s", fn)
		pprof.StartCPUProfile(f)
		prof.closer = func() {
			pprof.StopCPUProfile()
			f.Close()
			logf("xprof: cpu profiling disabled, %s", fn)
		}
	case modeMem:
		fn := filepath.Join(path, "mem.pprof")
		f, err := os.Create(fn)
		if err != nil {
			log.Fatalf("xprof: could not create memory profile %q: %v", fn, err)
		}
		old := runtime.MemProfileRate
		runtime.MemProfileRate = prof.memProfileRate
		logf("xprof: memory profiling enabled (rate %d), %s", runtime.MemProfileRate, fn)
		prof.closer = func() {
			pprof.Lookup(prof.memProfileType).WriteTo(f, 0)
			f.Close()
			runtime.MemProfileRate = old
			logf("xprof: memory profiling disabled, %s", fn)
		}
	case modeMutex:
		fn := filepath.Join(path, "mutex.pprof")
		f, err := os.Create(fn)
		if err != nil {
			log.Fatalf("xprof: could not create mutex profile %q: %v", fn, err)
		}
		runtime.SetMutexProfileFraction(1)
		logf("xprof: mutex profiling enabled, %s", fn)
		prof.closer = func() {
			if mp := pprof.Lookup("mutex"); mp != nil {
				mp.WriteTo(f, 0)
			}
			f.Close()
			runtime.SetMutexProfileFraction(0)
			logf("xprof: mutex profiling disabled, %s", fn)
		}
	case modeBlock:
		fn := filepath.Join(path, "block.pprof")
		f, err := os.Create(fn)
		if err != nil {
			log.Fatalf("xprof: could not create block profile %q: %v", fn, err)
		}
		runtime.SetBlockProfileRate(1)
		logf("xprof: block profiling enabled, %s", fn)
		prof.closer = func() {
			pprof.Lookup("block").WriteTo(f, 0)
			f.Close()
			runtime.SetBlockProfileRate(0)
			logf("xprof: block profiling disabled, %s", fn)
		}
	case modeThreadCreate:
		fn := filepath.Join(path, "threadcreation.pprof")
		f, err := os.Create(fn)
		if err != nil {
			log.Fatalf("xprof: could not create thread creation profile %q: %v", fn, err)
		}
		logf("xprof: thread creation profiling enabled, %s", fn)
		prof.closer = func() {
			if mp := pprof.Lookup("threadcreate"); mp != nil {
				mp.WriteTo(f, 0)
			}
			f.Close()
			logf("xprof: thread creation profiling disabled, %s", fn)
		}

	case modeTrace:
		fn := filepath.Join(path, "trace.out")
		f, err := os.Create(fn)
		if err != nil {
			log.Fatalf("xprof: could not create trace output file %q: %v", fn, err)
		}
		if err := trace.Start(f); err != nil {
			log.Fatalf("xprof: could not start trace: %v", err)
		}
		logf("xprof: trace enabled, %s", fn)
		prof.closer = func() {
			trace.Stop()
			logf("xprof: trace disabled, %s", fn)
		}
	case modeGoroutine:
		fn := filepath.Join(path, "goroutine.pprof")
		f, err := os.Create(fn)
		if err != nil {
			log.Fatalf("xprof: could not create goroutine profile %q: %v", fn, err)
		}
		logf("xprof: goroutine profiling enabled, %s", fn)
		prof.closer = func() {
			if mp := pprof.Lookup("goroutine"); mp != nil {
				mp.WriteTo(f, 0)
			}
			f.Close()
			logf("xprof: goroutine profiling disabled, %s", fn)
		}
	}
	if !prof.noShutdownHook {
		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			<-c

			log.Println("xprof: caught interrupt, stopping profiles")
			prof.Stop()

			os.Exit(0)
		}()
	}
	return &prof
}
