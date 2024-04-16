package opencodegraph

import (
	"context"

	"github.com/sourcegraph/sourcegraph/schema"
)

type Provider interface {
	Name() string
	Capabilities(ctx context.Context, params schema.CapabilitiesParams) (*schema.CapabilitiesResult, error)
	Annotations(ctx context.Context, params schema.AnnotationsParams) (*schema.AnnotationsResult, error)
}

var providers []Provider

func RegisterProvider(provider Provider) {
	providers = append(providers, provider)
}

var AllProviders Provider = &multiProvider{providers: &providers}
