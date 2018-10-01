package graphqlbackend

import (
	"context"
	"errors"
)

// DotcomMutation is the implementation of the GraphQL type DotcomMutation. If it is not set at
// runtime, a "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var DotcomMutation DotcomMutationResolver

func (schemaResolver) Dotcom() (DotcomMutationResolver, error) {
	if DotcomMutation == nil {
		return nil, errors.New("dotcom is not implemented")
	}
	return DotcomMutation, nil
}

// DotcomMutationResolver is the API of the GraphQL type DotcomMutation.
type DotcomMutationResolver interface {
	GenerateSourcegraphLicenseKey(ctx context.Context, args *DotcomMutationGenerateSourcegraphLicenseKeyArgs) (string, error)
}

type DotcomMutationGenerateSourcegraphLicenseKeyArgs struct {
	Plan         string
	MaxUserCount *int32
	ExpiresAt    *int32
}
