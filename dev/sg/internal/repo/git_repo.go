package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/run"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitRepo struct {
	// Branch is the branch of the repository
	Branch string
	// Ref is the current commit of the repository
	Ref string
}

var ErrBranchNotFound = errors.New("branch not found")

type pushOpt func(args []string) []string

func withForceLease(args []string) []string {
	return append(args, "--force-with-lease")
}

func WithForce(args []string) []string {
	return append(args, "--force")
}

func WithPushRefSpec(ref, branch string) pushOpt {
	return func(args []string) []string {
		idx := -1
		for i, v := range args {
			if v == branch {
				idx = i
			}
		}

		if idx < 0 {
			return args
		}
		args[idx] = fmt.Sprintf("%s:refs/heads/%s", ref, branch)
		return args
	}
}

func NewGitRepo(branch, ref string) *GitRepo {
	return &GitRepo{Branch: branch, Ref: ref}
}

func (g *GitRepo) IsOutOfSync(ctx context.Context) (bool, error) {
	if ok, err := g.HasRemoteBranch(ctx); err != nil {
		return false, nil
	} else if !ok {
		return false, nil
	}

	return !g.HasRemoteCommit(ctx), nil
}
func (g *GitRepo) PushToRemote(ctx context.Context) (string, error) {
	ok, err := g.HasRemoteBranch(ctx)
	if err != nil {
		return "", err
	}

	if ok {
		// push with lease only works if the branch exists remotely
		return g.Push(ctx, withForceLease)
	}
	return g.Push(ctx, WithPushRefSpec(g.Ref, g.Branch))
}

func (g *GitRepo) ListChangedFiles(ctx context.Context) ([]string, error) {
	files, err := run.Cmd(ctx, "git diff --name-only").Run().Lines()
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (g *GitRepo) IsDirty(ctx context.Context) (bool, error) {
	files, err := g.ListChangedFiles(ctx)
	if err != nil {
		return false, err
	}
	return len(files) > 0, nil
}

func (g *GitRepo) Push(ctx context.Context, opts ...pushOpt) (string, error) {
	cmd := []string{"git", "push", "origin", g.Branch}
	for _, opt := range opts {
		cmd = opt(cmd)
	}
	return run.Cmd(ctx, cmd...).Run().String()
}

func (g *GitRepo) PushWithLease(ctx context.Context) (string, error) {
	return run.Cmd(ctx, "git", "push", "origin", g.Branch, "--force-with-lease").Run().String()
}

func (g *GitRepo) GetHeadCommit() string {
	return g.Ref
}

func (g *GitRepo) HasRemoteBranch(ctx context.Context) (bool, error) {
	return HasRemoteBranch(ctx, g.Branch)
}

func (g *GitRepo) HasRemoteCommit(ctx context.Context) bool {
	return HasCommit(ctx, g.Ref)
}

func (g *GitRepo) FetchOrigin(ctx context.Context) (string, error) {
	output, err := run.Cmd(ctx, "git", "fetch", "origin", g.Branch).Run().String()
	if err != nil {
		if strings.Contains(output, "couldn't find remote ref") {
			return "", ErrBranchNotFound
		}
		return "", err
	}
	return strings.TrimSpace(output), nil
}
