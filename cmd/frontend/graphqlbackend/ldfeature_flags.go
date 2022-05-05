package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/ldfeatureflag"
)

func (r *schemaResolver) AllEnabledFeatureFlags(ctx context.Context) []string {
	return ldfeatureflag.AllEnabledFeatureFlags(ctx, r.db)
}

func (r *schemaResolver) FeatureFlagEnabled(ctx context.Context, args *struct {
	Name         string
	DefaultValue bool
}) bool {
	flag := ldfeatureflag.FeatureFlag{Name: args.Name, DefaultValue: args.DefaultValue}

	return flag.IsEnabledFor(ctx, r.db)
}
