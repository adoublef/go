// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wait

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ForHTTP waits for an HTTP endpoint to become available within the specified timeout.
// It uses exponential backoff to retry requests until the endpoint responds successfully
// or the context is canceled.
func ForHTTP(ctx context.Context, timeout time.Duration, endpoint string, opts ...func(*http.Request)) error {
	return ForFunc(ctx, timeout, func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		for _, o := range opts {
			o(req)
		}
		res, err := (&http.Client{}).Do(req)
		if err != nil {
			return fmt.Errorf("failed making request: %w", err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return NotReady
		}
		return nil
	})
}
