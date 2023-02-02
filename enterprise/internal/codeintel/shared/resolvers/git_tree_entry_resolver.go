package sharedresolvers

import (
	"context"
	"io/fs"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/cloneurls"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GitTreeEntryResolver resolves an entry in a Git tree in a repository. The entry can be any Git
// object type that is valid in a tree.
//
// Prefer using the constructor, NewGitTreeEntryResolver.
type GitTreeEntryResolver struct {
	db     database.DB
	commit *GitCommitResolver

	contentOnce sync.Once
	content     []byte
	contentErr  error

	// stat is this tree entry's file info. Its Name method must return the full path relative to
	// the root, not the basename.
	stat fs.FileInfo

	logger log.Logger
}

func NewGitTreeEntryResolver(db database.DB, commit *GitCommitResolver, stat fs.FileInfo) *GitTreeEntryResolver {
	return &GitTreeEntryResolver{db: db, commit: commit, stat: stat, logger: log.Scoped("git tree entry resolver", "")}
}
func (r *GitTreeEntryResolver) Path() string { return r.stat.Name() }
func (r *GitTreeEntryResolver) Name() string { return path.Base(r.stat.Name()) }

func (r *GitTreeEntryResolver) ToGitTree() (resolverstubs.GitTreeEntryResolver, bool) {
	return r, r.IsDirectory()
}

func (r *GitTreeEntryResolver) ToGitBlob() (resolverstubs.GitTreeEntryResolver, bool) {
	return r, !r.IsDirectory()
}

// func (r *GitTreeEntryResolver) ToVirtualFile() (*virtualFileResolver, bool) { return nil, false }

func (r *GitTreeEntryResolver) ByteSize(ctx context.Context) (int32, error) {
	content, err := r.Content(ctx, &resolverstubs.GitTreeContentPageArgs{})
	if err != nil {
		return 0, err
	}
	return int32(len([]byte(content))), nil
}

func (r *GitTreeEntryResolver) Content(ctx context.Context, args *resolverstubs.GitTreeContentPageArgs) (string, error) {
	r.contentOnce.Do(func() {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		r.content, r.contentErr = gitserver.NewClient().ReadFile(
			ctx,
			authz.DefaultSubRepoPermsChecker,
			api.RepoName(r.commit.Repository().Name()),
			api.CommitID(r.commit.OID()),
			r.Path(),
		)
	})

	return string(r.content), r.contentErr
}

func (r *GitTreeEntryResolver) Commit() resolverstubs.GitCommitResolver {
	return r.commit
}

func (r *GitTreeEntryResolver) Repository() resolverstubs.RepositoryResolver {
	return r.commit.Repository()
}

func (r *GitTreeEntryResolver) CanonicalURL() string {
	canonicalURL := r.commit.canonicalRepoRevURL()
	return r.urlPath(canonicalURL).String()
}

func (r *GitTreeEntryResolver) IsRoot() bool {
	cleanedPath := path.Clean(r.Path())
	return cleanedPath == "/" || cleanedPath == "." || cleanedPath == ""
}

func (r *GitTreeEntryResolver) IsDirectory() bool { return r.stat.Mode().IsDir() }

func (r *GitTreeEntryResolver) URL(ctx context.Context) (string, error) {
	return r.url(ctx).String(), nil
}

func (r *GitTreeEntryResolver) Submodule() resolverstubs.GitSubmoduleResolver {
	if r == nil {
		r.logger.Error("git tree entry resolver is nil", log.Error(errors.New("git tree entry resolver is nil")))
		return nil
	}

	if r.stat == nil {
		r.logger.Error("stat is nil", log.Error(errors.New("stat is nil")))
		return nil
	}

	if submoduleInfo, ok := r.stat.Sys().(gitdomain.Submodule); ok {
		return NewGitSubmoduleResolver(submoduleInfo)
	}
	return nil
}

func (r *GitTreeEntryResolver) url(ctx context.Context) *url.URL {
	if submodule := r.Submodule(); submodule != nil {
		submoduleURL := submodule.URL()
		if strings.HasPrefix(submoduleURL, "../") {
			submoduleURL = path.Join(r.Repository().Name(), submoduleURL)
		}
		repoName, err := cloneURLToRepoName(ctx, r.db, submoduleURL)
		if err != nil {
			log.Error(err)
			return &url.URL{}
		}
		return &url.URL{Path: "/" + repoName + "@" + submodule.Commit()}
	}
	return r.urlPath(r.commit.repoRevURL())
}

func (r *GitTreeEntryResolver) urlPath(prefix *url.URL) *url.URL {
	// Dereference to copy to avoid mutating the input
	u := *prefix
	if r.IsRoot() {
		return &u
	}

	typ := "blob"
	if r.IsDirectory() {
		typ = "tree"
	}

	u.Path = path.Join(u.Path, "-", typ, r.Path())
	return &u
}

func cloneURLToRepoName(ctx context.Context, db database.DB, cloneURL string) (string, error) {
	repoName, err := cloneurls.RepoSourceCloneURLToRepoName(ctx, db, cloneURL)
	if err != nil {
		return "", err
	}
	if repoName == "" {
		return "", errors.Errorf("no matching code host found for %s", cloneURL)
	}
	return string(repoName), nil
}
