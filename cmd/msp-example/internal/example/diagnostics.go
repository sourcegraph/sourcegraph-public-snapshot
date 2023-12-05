package example

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

type serviceState struct {
	statelessMode bool
	contract      runtime.Contract
}

func (s serviceState) Healthy(ctx context.Context) error {
	if s.statelessMode {
		return nil
	}

	// Write a single test event
	if err := writeBigQueryEvent(ctx, s.contract, "service.healthy"); err != nil {
		return errors.Wrap(err, "writeBigQueryEvent")
	}

	return nil
}
