package service

import (
	"context"
	"net/url"
)

type serviceState struct{}

func (s serviceState) Healthy(ctx context.Context, query url.Values) error {
	return nil
}
