package adapters

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/domain"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

type Git struct {
	ReposDir string // The root directory where repos are stored
}

func (g *Git) RevParse(ctx context.Context, repo api.RepoName, rev string) (string, error) {
	return "", errors.New("TODO")
}

// GetObjectType returns the object type given an objectID
func (g *Git) GetObjectType(ctx context.Context, repo api.RepoName, objectID string) (domain.ObjectType, error) {
	cmd := exec.CommandContext(ctx, "git", "cat-file", "-t", "--", objectID)
	cmd.Path = repoDir(repo, g.ReposDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args, out))
	}

	objectType := domain.ObjectType(bytes.TrimSpace(out))
	return objectType, nil
}

func repoDir(name api.RepoName, reposDir string) string {
	path := string(protocol.NormalizeRepo(name))
	return filepath.Join(reposDir, filepath.FromSlash(path), ".git")
}
