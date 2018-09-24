package graphqlbackend

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/licensing"
)

func (schemaResolver) Dotcom() *dotcomMutationResolver { return &dotcomMutationResolver{} }

// dotcomMutationResolver implements the GraphQL type DotcomMutation.
type dotcomMutationResolver struct{}

func (dotcomMutationResolver) GenerateSourcegraphLicenseKey(ctx context.Context, args *licensing.GenerateSourcegraphLicenseKeyArgs) (string, error) {
	// 🚨 SECURITY: Only site admins may generate Sourcegraph license keys.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return "", err
	}

	if f := licensing.GenerateSourcegraphLicenseKey; f != nil {
		return f(ctx, args)
	}
	return "", errors.New("not implemented")
}
