package sharedresolvers

import (
	"context"
	"io/fs"
	"net/url"
	"path"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database"
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

	// stat is this tree entry's file info. Its Name method must return the full path relative to
	// the root, not the basename.
	stat fs.FileInfo
}

func NewGitTreeEntryResolver(db database.DB, commit *GitCommitResolver, stat fs.FileInfo) *GitTreeEntryResolver {
	return &GitTreeEntryResolver{db: db, commit: commit, stat: stat}
}
func (r *GitTreeEntryResolver) Path() string { return r.stat.Name() }
func (r *GitTreeEntryResolver) Name() string { return path.Base(r.stat.Name()) }

func (r *GitTreeEntryResolver) Commit() *GitCommitResolver      { return r.commit }
func (r *GitTreeEntryResolver) Repository() *RepositoryResolver { return r.commit.repoResolver }
func (r *GitTreeEntryResolver) CanonicalURL() string {
	url := r.commit.canonicalRepoRevURL()
	return r.urlPath(url).String()
}

func (r *GitTreeEntryResolver) IsRoot() bool {
	path := path.Clean(r.Path())
	return path == "/" || path == "." || path == ""
}

func (r *GitTreeEntryResolver) IsDirectory() bool { return r.stat.Mode().IsDir() }

func (r *GitTreeEntryResolver) URL(ctx context.Context) (string, error) {
	return r.url(ctx).String(), nil
}

func (r *GitTreeEntryResolver) Submodule() *gitSubmoduleResolver {
	if submoduleInfo, ok := r.stat.Sys().(gitdomain.Submodule); ok {
		// return &gitSubmoduleResolver{submodule: submoduleInfo}
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
			log15.Error("Failed to resolve submodule repository name from clone URL", "cloneURL", submodule.URL(), "err", err)
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
	repoName, err := ReposourceCloneURLToRepoName(ctx, db, cloneURL)
	if err != nil {
		return "", err
	}
	if repoName == "" {
		return "", errors.Errorf("no matching code host found for %s", cloneURL)
	}
	return string(repoName), nil
}
