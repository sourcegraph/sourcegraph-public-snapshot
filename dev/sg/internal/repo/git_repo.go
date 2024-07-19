package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ErrBranchNotFound is returned when a branch is not found
var ErrBranchNotFound = errors.New("branch not found")

// pushOpt is a function that modifies the arguments to git push
type pushOpt func(args []string) []string

// GitRepo represents a branch at a commit of a git repository. No assumption is made on what repository is operated on.
//
// Various methods to operate on a branch locally and with its remote state in a git repository which may or may not be the sourcegraph repository.
//
// In contrast, the State struct just considers the state of the repository in reference to its merge base and how the current state of the repository differs with its merge base.
type GitRepo struct {
	// Branch is the branch of the repository
	Branch string
	// Ref is the current commit of the repository
	Ref string
}

// NewGitRepo returns a new GitRepo with the given branch and commit
func NewGitRepo(branch, ref string) *GitRepo {
	return &GitRepo{Branch: branch, Ref: ref}
}

// WithForceLease adds --force-with-lease to the args. The provided args are arguments a git push command
func WithForceLease(args []string) []string {
	return append(args, "--force-with-lease")
}

// WithPushRefSpec adds a refspec to the args. If the it replaces the branch argument if it is already specified.
// The provided args are arguments for a git push command.
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

// IsOutOfSync checks whether the remote state of this branch is in sync with the local state of this branch
// The state is determined by checking:
// * Does the branch exist remotely?
// * Does the commit exist remotely?
func (g *GitRepo) IsOutOfSync(ctx context.Context) (bool, error) {
	if ok, err := g.HasRemoteBranch(ctx); err != nil {
		println("HAS REMOTE ERR")
		println("HAS REMOTE ERR")
		println("HAS REMOTE ERR")
		return false, err
	} else if !ok {
		// We don't have a remote branch, so we're definitely out of sync
		return true, nil
	}

	// Now lets check if the commit exists in the remote branch
	return !g.HasRemoteCommit(ctx), nil
}

func (g *GitRepo) checkout(ctx context.Context, args ...string) error {
	checkoutArgs := []string{"git", "checkout"}
	checkoutArgs = append(checkoutArgs, args...)
	err := run.Cmd(ctx, checkoutArgs...).Run().Wait()
	return err
}

func (g *GitRepo) Checkout(ctx context.Context) error {
	return g.checkout(ctx, g.Branch)
}

func (g *GitRepo) CheckoutNewBranch(ctx context.Context) error {
	return g.checkout(ctx, "-b", g.Branch)
}

// ListChangedFiles lists the files that have changed since the last commit
func (g *GitRepo) ListChangedFiles(ctx context.Context) ([]string, error) {
	files, err := run.Cmd(ctx, "git diff --name-only").Run().Lines()
	if err != nil {
		return nil, err
	}

	return files, nil
}

// IsDirty checks whether the current repository state has any uncommited changes
func (g *GitRepo) IsDirty(ctx context.Context) (bool, error) {
	files, err := g.ListChangedFiles(ctx)
	if err != nil {
		return false, err
	}
	return len(files) > 0, nil
}

// Add a file to staged changes
func (g *GitRepo) Add(ctx context.Context, file string) (string, error) {
	return run.Cmd(ctx, "git", "add", file).Run().String()
}

// Commit commits the staged changes with the given message
func (g *GitRepo) Commit(ctx context.Context, message string) (string, error) {
	return run.Cmd(ctx, "git", "commit", "-m", message).Run().String()
}

// Push pushes the current branch to origin using the specific push options
func (g *GitRepo) Push(ctx context.Context, opts ...pushOpt) (string, error) {
	cmd := []string{"git", "push", "origin", g.Branch}
	for _, opt := range opts {
		cmd = opt(cmd)
	}
	return run.Cmd(ctx, cmd...).Run().String()
}

// GetHeadCommit returns the current ref that is considered to be the HEAD
func (g *GitRepo) GetHeadCommit() string {
	return g.Ref
}

// HasLocalBranch checks whether the current branch exists locally
func (g *GitRepo) HasLocalBranch(ctx context.Context) (bool, error) {
	result, err := run.Cmd(ctx, "git", "branch", "--list", g.Branch).Run().String()
	if err != nil {
		return false, err
	}

	result = strings.TrimSpace(result)
	return len(result) > 0, nil

}

// HasRemoteBranch checks whether the current branch exists remotely
func (g *GitRepo) HasRemoteBranch(ctx context.Context) (bool, error) {
	return HasRemoteBranch(ctx, g.Branch)
}

// HasRemoteCommit checks whether the current commit exists remotely
func (g *GitRepo) HasRemoteCommit(ctx context.Context) bool {
	return HasCommit(ctx, g.Ref)
}

// FetchOrigin fetches the current branch from origin. If the branch ref doesn't exist remotely ErrBranchNotFound is returned
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
