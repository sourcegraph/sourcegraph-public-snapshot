package adapters

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Git struct {
	// ReposDir is the root directory where repos are stored.
	ReposDir string
	// RecordingCommandFactory is a factory that creates recordable commands by
	// wrapping os/exec.Commands. The factory creates recordable commands with a set
	// predicate, which is used to determine whether a particular command should be
	// recorded or not.
	RecordingCommandFactory *wrexec.RecordingCommandFactory
}

// RevParse will run rev-parse on the given rev
func (g *Git) RevParse(ctx context.Context, repo api.RepoName, rev string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", rev)
	cmd.Dir = repoDir(repo, g.ReposDir)
	wrappedCmd := g.RecordingCommandFactory.WrapWithRepoName(ctx, log.NoOp(), repo, cmd)
	out, err := wrappedCmd.CombinedOutput()
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", wrappedCmd.Args, out))
	}

	return string(out), nil
}

// GetObjectType returns the object type given an objectID
func (g *Git) GetObjectType(ctx context.Context, repo api.RepoName, objectID string) (gitdomain.ObjectType, error) {
	cmd := exec.CommandContext(ctx, "git", "cat-file", "-t", "--", objectID)
	cmd.Dir = repoDir(repo, g.ReposDir)
	wrappedCmd := g.RecordingCommandFactory.WrapWithRepoName(ctx, log.NoOp(), repo, cmd)
	out, err := wrappedCmd.CombinedOutput()
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", wrappedCmd.Args, out))
	}

	objectType := gitdomain.ObjectType(bytes.TrimSpace(out))
	return objectType, nil
}

func repoDir(name api.RepoName, reposDir string) string {
	path := string(protocol.NormalizeRepo(name))
	return filepath.Join(reposDir, filepath.FromSlash(path), ".git")
}
