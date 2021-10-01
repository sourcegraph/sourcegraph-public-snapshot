package adapters

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/domain"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type Git struct{}

func (g *Git) RevParse(ctx context.Context, repo api.RepoName, rev string) (string, error) {
	return "", errors.New("TODO")
}

func (g *Git) GetObjectType(ctx context.Context, repo api.RepoName, rev string) (*domain.GitObject, error) {
	return nil, errors.New("TODO")
}
