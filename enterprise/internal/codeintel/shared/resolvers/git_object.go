package sharedresolvers

import (
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitObjectType string

func (GitObjectType) ImplementsGraphQLType(name string) bool { return name == "GitObjectType" }

const (
	GitObjectTypeCommit  GitObjectType = "GIT_COMMIT"
	GitObjectTypeTag     GitObjectType = "GIT_TAG"
	GitObjectTypeTree    GitObjectType = "GIT_TREE"
	GitObjectTypeBlob    GitObjectType = "GIT_BLOB"
	GitObjectTypeUnknown GitObjectType = "GIT_UNKNOWN"
)

type GitObjectID string

func (GitObjectID) ImplementsGraphQLType(name string) bool {
	return name == "GitObjectID"
}

func (id *GitObjectID) UnmarshalGraphQL(input any) error {
	if input, ok := input.(string); ok && gitserver.IsAbsoluteRevision(input) {
		*id = GitObjectID(input)
		return nil
	}
	return errors.New("GitObjectID: expected 40-character string (SHA-1 hash)")
}
