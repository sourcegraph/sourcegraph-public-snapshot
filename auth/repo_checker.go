package auth

import (
	"os"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// A RepoChecker checks the actor's permissions on a repo.
type RepoChecker interface {
	// CheckRepo checks whether the actor has permission to take the
	// specified kind of action on the repository.
	//
	// If the action is authorized, it should nil. Otherwise it must
	// return a non-nil error (os.ErrPermission in the general
	// case).
	CheckRepo(ctx context.Context, repo string, what PermType) error
}

// RepoCheckerFunc implements RepoChecker by calling itself.
type RepoCheckerFunc func(ctx context.Context, repo string, what PermType) error

// CheckRepo implements RepoChecker.
func (f RepoCheckerFunc) CheckRepo(ctx context.Context, repo string, what PermType) error {
	return f(ctx, repo, what)
}

// RepoPermsFunc implements RepoChecker by calling itself to get
// the actor's repo permissions and then checking them against the
// requested permission.
type RepoPermsFunc func(ctx context.Context, repo string) (*sourcegraph.RepoPermissions, error)

// CheckRepo implements RepoChecker.
func (f RepoPermsFunc) CheckRepo(ctx context.Context, repo string, perm PermType) error {
	perms, err := f(ctx, repo)
	if err != nil {
		return err
	}
	if perms == nil {
		panic("perms == nil")
	}

	var hasPerm bool
	switch perm {
	case Read:
		hasPerm = perms.Read || perms.Write || perms.Admin
	case Write:
		hasPerm = perms.Write || perms.Admin
	case Admin:
		hasPerm = perms.Admin
	default:
		panic("unrecognized perm")
	}

	if !hasPerm {
		return os.ErrPermission
	}
	return nil
}

// WithRepoChecker creates a child context that uses the specified
// func to check whether the actor has the specified permission on a
// repository.
func WithRepoChecker(ctx context.Context, rc RepoChecker) context.Context {
	if rc == nil {
		panic("RepoChecker is nil")
	}
	return context.WithValue(ctx, repoCheckerKey, rc)
}

// CheckRepo checks whether the context's actor has the specified
// permission on a repository.
//
// If the action is authorized, it returns nil. Otherwise it returns a
// non-nil error (os.ErrPermission in the general case).
func CheckRepo(ctx context.Context, repo string, what PermType) error {
	if ctx == nil {
		panic("ctx == nil")
	}
	if _, ok := ctx.Value(repoCheckerStartedKey).(struct{}); ok {
		panic("CheckRepo called recursively")
	}
	rc, _ := ctx.Value(repoCheckerKey).(RepoChecker)
	if rc == nil {
		panic("no RepoChecker set in context")
	}

	// Mark the context so that we can detect recursion.
	ctx = context.WithValue(ctx, repoCheckerStartedKey, struct{}{})

	return rc.CheckRepo(ctx, repo, what)
}
