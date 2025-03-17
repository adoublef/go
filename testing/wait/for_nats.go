// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wait

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/nats-io/nats-server/v2/server"
)

// ForNATS
func ForNATS(ctx context.Context, ns *server.Server, timeout time.Duration) error {
	o := func() error {
		if !ns.ReadyForConnections(10 * time.Millisecond) {
			return fmt.Errorf("nats server not ready")
		}
		return nil
	}
	bo := backoff.NewExponentialBackOff(backoff.WithMaxElapsedTime(timeout))
	return backoff.Retry(o, backoff.WithContext(bo, ctx))
}
