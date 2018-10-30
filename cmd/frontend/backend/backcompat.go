package backend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// TODO!(sqs): This file contains backcompat stubs for definitions that were removed in the
// migration to using Sourcegraph extensions for language support.

var MockBackcompatBackendDefsTotalRefs func(ctx context.Context, repo api.RepoURI) (int, error)

func BackcompatBackendDefsTotalRefs(ctx context.Context, repo api.RepoURI) (int, error) {
	if MockBackcompatBackendDefsTotalRefs != nil {
		return MockBackcompatBackendDefsTotalRefs(ctx, repo)
	}
	panic("TODO!(sqs): removed")
}
