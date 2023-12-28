package example

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

type serviceState struct {
	statelessMode bool
	contract      runtime.Contract
}

func (s serviceState) Healthy(ctx context.Context, _ url.Values) error {
	if s.statelessMode {
		return nil
	}

	// Write a single test event
	if err := writeBigQueryEvent(ctx, s.contract, "service.healthy"); err != nil {
		return errors.Wrap(err, "writeBigQueryEvent")
	}

	// Check redis connection
	if _, err := newRedisConnection(ctx, s.contract); err != nil {
		return errors.Wrap(err, "newRedisConnection")
	}

	return nil
}
