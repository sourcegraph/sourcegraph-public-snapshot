package sharedresolvers

import (
	"net/url"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type gitCommitResolver struct {
	// oid MUST be specified and a 40-character Git SHA.
	oid gitObjectID

	repoResolver *RepositoryResolver

	// inputRev is the Git revspec that the user originally requested that resolved to this Git commit. It is used
	// to avoid redirecting a user browsing a revision "mybranch" to the absolute commit ID as they follow links in the UI.
	inputRev *string
}

// newGitCommitResolver returns a new CommitResolver. When commit is set to nil,
// commit will be loaded lazily as needed by the resolver. Pass in a commit when
// you have batch-loaded a bunch of them and already have them at hand.
func newGitCommitResolver(repo *RepositoryResolver, id api.CommitID) *gitCommitResolver {
	return &gitCommitResolver{
		oid:          gitObjectID(id),
		repoResolver: repo,
	}
}

func (r *gitCommitResolver) ID() graphql.ID {
	return resolverstubs.MarshalID("GitCommit", struct {
		Repository graphql.ID  `json:"r"`
		CommitID   gitObjectID `json:"c"`
	}{Repository: r.repoResolver.ID(), CommitID: r.oid})
}

func (r *gitCommitResolver) Repository() resolverstubs.RepositoryResolver { return r.repoResolver }

func (r *gitCommitResolver) OID() resolverstubs.GitObjectID { return resolverstubs.GitObjectID(r.oid) }

func (r *gitCommitResolver) AbbreviatedOID() string {
	return string(r.oid)[:7]
}

func (r *gitCommitResolver) URL() string {
	u := r.repoResolver.url()
	u.Path += "/-/commit/" + r.inputRevOrImmutableRev()
	return u.String()
}

// inputRevOrImmutableRev returns the input revspec, if it is provided and nonempty. Otherwise it returns the
// canonical OID for the revision.
func (r *gitCommitResolver) inputRevOrImmutableRev() string {
	if r.inputRev != nil && *r.inputRev != "" {
		return *r.inputRev
	}
	return string(r.oid)
}

func (r *gitCommitResolver) canonicalRepoRevURL() *url.URL {
	return &url.URL{Path: "/" + r.repoResolver.Name() + "@" + string(r.oid)}
}

// repoRevURL returns the URL path prefix to use when constructing URLs to resources at this
// revision. Unlike inputRevOrImmutableRev, it does NOT use the OID if no input revspec is
// given. This is because the convention in the frontend is for repo-rev URLs to omit the "@rev"
// portion (unlike for commit page URLs, which must include some revspec in
// "/REPO/-/commit/REVSPEC").
func (r *gitCommitResolver) repoRevURL() *url.URL {
	repoURL := &url.URL{Path: "/" + r.repoResolver.Name()}
	var rev string
	if r.inputRev != nil {
		rev = *r.inputRev // use the original input rev from the user
	} else {
		rev = string(r.oid)
	}
	if rev != "" {
		repoURL.Path += "@" + rev
	}
	return repoURL
}

type gitObjectID string

func (id *gitObjectID) UnmarshalGraphQL(input any) error {
	if input, ok := input.(string); ok && gitserver.IsAbsoluteRevision(input) {
		*id = gitObjectID(input)
		return nil
	}
	return errors.New("GitObjectID: expected 40-character string (SHA-1 hash)")
}
