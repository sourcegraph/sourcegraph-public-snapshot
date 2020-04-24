package server

import (
	"bytes"
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func getTipCommit(repositoryID int) (string, error) {
	repo, err := db.Repos.Get(context.Background(), api.RepoID(repositoryID))
	if err != nil {
		return "", err
	}

	cmd := gitserver.DefaultClient.Command("git", "rev-parse", "HEAD")
	cmd.Repo = gitserver.Repo{Name: repo.Name}
	out, err := cmd.CombinedOutput(context.Background())
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(out)), nil
}
