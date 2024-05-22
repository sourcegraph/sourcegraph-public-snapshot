package service

import (
	"context"
	"net/url"
)

type serviceState struct{}

func (s serviceState) Healthy(_ context.Context, _ url.Values) error { return nil }
