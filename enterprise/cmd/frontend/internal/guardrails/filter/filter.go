package filter

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/guardrails/attribution"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// ServiceOpts configures Service.
type ServiceOpts struct {
	// Attribution is used to search for matches.
	Attribution *attribution.Service

	// Config is used to query how filtering is configured on the instance.
	Config conftypes.SiteConfigQuerier
}

// Service is for the filter service which searches for any matches on
// snippets of code.
//
// Use NewService to construct this value.
type Service struct {
	ServiceOpts
}

type FilterConfiguration struct {
}

// NewService returns a service configured with observationCtx.
//
// Note: this registers metrics so should only be called once with the same
// observationCtx.
func NewService(observationCtx *observation.Context, opts ServiceOpts) *Service {
	return &Service{
		operations:  newOperations(observationCtx),
		ServiceOpts: opts,
	}
}
