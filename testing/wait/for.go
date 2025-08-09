// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wait

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// ForFunc waits for function to return non-nil error within the specified timeout.
// It uses exponential backoff to retry requests until the endpoint responds successfully
// or the context is canceled.
func ForFunc(ctx context.Context, timeout time.Duration, f func() error) error {
	o := func() error {
		if err := f(); err != nil {
			if err == SkipRetry {
				return &backoff.PermanentError{Err: SkipRetry}
			}
			return err
		}
		return nil
	}
	bo := backoff.NewExponentialBackOff(backoff.WithMaxElapsedTime(timeout))
	return backoff.Retry(o, backoff.WithContext(bo, ctx))
}

// SkipRetry is used as a return value from [ForFunc]
var SkipRetry = errors.New("skip retry")
