package graphqlbackend

import (
	"context"
	"time"

	"github.com/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	"github.com/sourcegraph/enterprise/pkg/license"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func init() {
	// Contribute the GraphQL type DotcomMutation.
	graphqlbackend.DotcomMutation = dotcomMutationResolver{}
}

// dotcomMutationResolver implements the GraphQL type DotcomMutation.
type dotcomMutationResolver struct{}

func (dotcomMutationResolver) GenerateSourcegraphLicenseKey(ctx context.Context, args *graphqlbackend.DotcomMutationGenerateSourcegraphLicenseKeyArgs) (string, error) {
	// 🚨 SECURITY: Only site admins may generate Sourcegraph license keys.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return "", err
	}

	info := license.Info{Plan: args.Plan}
	if args.MaxUserCount != nil {
		n := uint(*args.MaxUserCount)
		info.MaxUserCount = &n
	}
	if args.ExpiresAt != nil {
		t := time.Unix(int64(*args.ExpiresAt), 0)
		info.Expiry = &t
	}
	return licensing.GenerateSourcegraphLicenseKey(info)
}
