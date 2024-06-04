package service

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type serviceState struct {
	dotcomDB interface{ Ping(context.Context) error }
}

func (s serviceState) Healthy(ctx context.Context, query url.Values) error {
	if query.Has("dotcomdb") {
		if err := s.dotcomDB.Ping(ctx); err != nil {
			return errors.Wrap(err, "dotcomdb.Ping")
		}
	}
	return nil
}
