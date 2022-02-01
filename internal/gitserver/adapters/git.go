package adapters

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

type Git struct {
	ReposDir string // The root directory where repos are stored
}

// RevParse will run rev-parse on the given rev
func (g *Git) RevParse(ctx context.Context, repo api.RepoName, rev string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", rev)
	cmd.Dir = repoDir(repo, g.ReposDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args, out))
	}

	return string(out), nil
}

// GetObjectType returns the object type given an objectID
func (g *Git) GetObjectType(ctx context.Context, repo api.RepoName, objectID string) (gitdomain.ObjectType, error) {
	cmd := exec.CommandContext(ctx, "git", "cat-file", "-t", "--", objectID)
	cmd.Dir = repoDir(repo, g.ReposDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args, out))
	}

	objectType := gitdomain.ObjectType(bytes.TrimSpace(out))
	return objectType, nil
}

func repoDir(name api.RepoName, reposDir string) string {
	path := string(protocol.NormalizeRepo(name))
	return filepath.Join(reposDir, filepath.FromSlash(path), ".git")
}
