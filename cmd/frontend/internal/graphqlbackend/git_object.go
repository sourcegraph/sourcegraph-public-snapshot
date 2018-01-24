package graphqlbackend

import (
	"context"
	"errors"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
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
func (o *gitObject) AbbreviatedOID(ctx context.Context) (string, error) {
	return string(o.oid[:7]), nil
}

type gitObjectResolver struct {
	repo    *repositoryResolver
	revspec string

	once sync.Once
	oid  gitObjectID
	err  error
}

func (o *gitObjectResolver) resolve(ctx context.Context) (gitObjectID, error) {
	o.once.Do(func() {
		resolvedRev, err := backend.Repos.ResolveRev(ctx, o.repo.repo.ID, o.revspec)
		if err != nil {
			o.err = err
			return
		}
		o.oid = gitObjectID(resolvedRev)
	})
	return o.oid, o.err
}

func (o *gitObjectResolver) OID(ctx context.Context) (gitObjectID, error) {
	return o.resolve(ctx)
}

func (o *gitObjectResolver) AbbreviatedOID(ctx context.Context) (string, error) {
	oid, err := o.resolve(ctx)
	if err != nil {
		return "", err
	}
	return string(oid[:7]), nil
}
