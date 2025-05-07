// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"go.adoublef.dev/runtime/debug"
)

// Decode reads the next JSON-encoded value from a [http.Request] and returns the value is valid.
//
// If sz or d is set, the max bytes and read deadline of the [http.Request] can be modified, respectively.
func Decode[V any](w http.ResponseWriter, r *http.Request, sz int, d time.Duration) (V, error) {
	var v V
	if r.Body == nil {
		return v, newDecodeErr(http.StatusUnauthorized, "request body could not be read properly")
	}
	mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || !(mt == "application/json") {
		return v, newDecodeErr(http.StatusUnsupportedMediaType, "request body could not be read properly")
	}
	if sz > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, int64(sz))
		debug.Printf("r.Body = http.MaxBytesReader(w, r.Body, %d)", sz)
	}

	if d > 0 {
		rc := http.NewResponseController(w)
		err = rc.SetReadDeadline(time.Now().Add(d))
		debug.Printf("%s := rc.SetReadDeadline(time.Now().Add(%v))", err, d)
		if err != nil {
			// note: if action not allowed, should maybe wrap this
			return v, err
		}
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // important
	if err := dec.Decode(&v); err != nil {
		debug.Printf("%v := dec.Decode(&v)", err)
		var zero V
		switch {
		// In some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON. There
		// is an open issue regarding this at
		// https://github.com/golang/go/issues/25956.
		case errors.As(err, new(*json.SyntaxError)):
			se := err.(*json.SyntaxError)
			ch, _ := utf8.DecodeRune([]byte(se.Error()[19:]))
			return zero, newDecodeErr(http.StatusBadRequest, "invalid character '%c' at position %d", ch, se.Offset)
		case errors.As(err, new(*json.UnmarshalTypeError)):
			e := err.(*json.UnmarshalTypeError)
			return zero, newDecodeErr(http.StatusBadRequest, "unexpected %s for field %q at position %d", e.Value, e.Field, e.Offset)
		// There is an open issue at https://github.com/golang/go/issues/29035
		// regarding turning this into a sentinel error.
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			return zero, newDecodeErr(http.StatusBadRequest, "unknown field %s", err.Error()[20:])
		// An io.EOF error is returned by Decode() if the request body is empty.
		case errors.Is(err, io.EOF):
			return zero, newDecodeErr(http.StatusUnauthorized, "request body could not be read properly")
		case errors.As(err, new(*http.MaxBytesError)):
			return zero, newDecodeErr(http.StatusRequestEntityTooLarge, "maximum allowed request size is %d", sz)
		case errors.As(err, new(*net.OpError)):
			return zero, newDecodeErr(http.StatusRequestTimeout, "failed to process request in time, please try again")

		// Otherwise default to logging the error and sending a 500 Internal
		// Server Error response. May want to wrap this error.
		default:
			return zero, newDecodeErr(http.StatusBadRequest, "encoding error: %v", err)
		}
	}
	// note: log error as this will not be returned to the client
	// Call decode again, using a pointer to an empty anonymous struct as
	// the destination. If the request body only contained a single JSON
	// object this will return an io.EOF error. So if we get anything else,
	// we know that there is additional data in the request body.
	if err = dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		// fixme: 4xx
		return *new(V), newDecodeErr(http.StatusBadRequest, "request body contains more than a single JSON object")
	}
	return v, nil
}

var ErrDecodeBody = errors.New("invalid decode")

type DecodeError struct {
	Code   int
	Reason string
	Err    error
}

func (e DecodeError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Reason)
}

func (e DecodeError) Is(target error) bool { return target == ErrDecodeBody }

func (e *DecodeError) Unwrap() error { return e.Err }

func newDecodeErr(code int, format string, v ...any) error {
	return &DecodeError{Code: code, Reason: fmt.Sprintf(format, v...)}
}
