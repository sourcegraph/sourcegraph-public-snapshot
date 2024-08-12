package http

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type RepoStore interface {
	GetByName(ctx context.Context, name api.RepoName) (*types.Repo, error)
}
