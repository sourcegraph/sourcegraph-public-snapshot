package sharedresolvers

import (
	"net/url"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type GitCommitResolver struct {
	// oid MUST be specified and a 40-character Git SHA.
	oid GitObjectID

	repoResolver *RepositoryResolver

	// inputRev is the Git revspec that the user originally requested that resolved to this Git commit. It is used
	// to avoid redirecting a user browsing a revision "mybranch" to the absolute commit ID as they follow links in the UI.
	inputRev *string
}

// NewGitCommitResolver returns a new CommitResolver. When commit is set to nil,
// commit will be loaded lazily as needed by the resolver. Pass in a commit when
// you have batch-loaded a bunch of them and already have them at hand.
func NewGitCommitResolver(repo *RepositoryResolver, id api.CommitID) *GitCommitResolver {
	return &GitCommitResolver{
		oid:          GitObjectID(id),
		repoResolver: repo,
	}
}

func (r *GitCommitResolver) ID() graphql.ID {
	return marshalGitCommitID(r.repoResolver.ID(), r.oid)
}

func (r *GitCommitResolver) Repository() resolverstubs.RepositoryResolver { return r.repoResolver }

func (r *GitCommitResolver) OID() resolverstubs.GitObjectID { return resolverstubs.GitObjectID(r.oid) }

func (r *GitCommitResolver) AbbreviatedOID() string {
	return string(r.oid)[:7]
}

func (r *GitCommitResolver) URL() string {
	u := r.repoResolver.url()
	u.Path += "/-/commit/" + r.inputRevOrImmutableRev()
	return u.String()
}

// inputRevOrImmutableRev returns the input revspec, if it is provided and nonempty. Otherwise it returns the
// canonical OID for the revision.
func (r *GitCommitResolver) inputRevOrImmutableRev() string {
	if r.inputRev != nil && *r.inputRev != "" {
		return *r.inputRev
	}
	return string(r.oid)
}

func (r *GitCommitResolver) canonicalRepoRevURL() *url.URL {
	// Dereference to copy the URL to avoid mutation
	repoURL := *r.repoResolver.RepoMatch.URL()
	repoURL.Path += "@" + string(r.oid)
	return &repoURL
}

// repoRevURL returns the URL path prefix to use when constructing URLs to resources at this
// revision. Unlike inputRevOrImmutableRev, it does NOT use the OID if no input revspec is
// given. This is because the convention in the frontend is for repo-rev URLs to omit the "@rev"
// portion (unlike for commit page URLs, which must include some revspec in
// "/REPO/-/commit/REVSPEC").
func (r *GitCommitResolver) repoRevURL() *url.URL {
	// Dereference to copy to avoid mutation
	repoURL := *r.repoResolver.RepoMatch.URL()
	var rev string
	if r.inputRev != nil {
		rev = *r.inputRev // use the original input rev from the user
	} else {
		rev = string(r.oid)
	}
	if rev != "" {
		repoURL.Path += "@" + rev
	}
	return &repoURL
}
