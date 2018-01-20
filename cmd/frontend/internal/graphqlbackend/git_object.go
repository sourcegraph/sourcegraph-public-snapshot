package graphqlbackend

import (
	"context"
	"errors"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type gitObjectID string

func (gitObjectID) ImplementsGraphQLType(name string) bool {
	return name == "GitObjectID"
}

func (id *gitObjectID) UnmarshalGraphQL(input interface{}) error {
	if input, ok := input.(string); ok && isValidGitObjectID(input) {
		*id = gitObjectID(input)
		return nil
	}
	return errors.New("GitObjectID: expected 40-character string (SHA-1 hash)")
}

func isValidGitObjectID(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

type gitObject struct {
	oid gitObjectID
}

func (o *gitObject) OID(ctx context.Context) (gitObjectID, error) { return o.oid, nil }

type gitObjectResolver struct {
	repo    *repositoryResolver
	revspec string
}

func (o *gitObjectResolver) OID(ctx context.Context) (gitObjectID, error) {
	resolvedRev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: o.repo.repo.ID, Rev: o.revspec})
	if err != nil {
		return "", err
	}
	return gitObjectID(resolvedRev.CommitID), nil
}
