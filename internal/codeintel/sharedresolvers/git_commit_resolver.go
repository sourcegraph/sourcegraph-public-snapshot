package sharedresolvers

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
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

func (r *GitCommitResolver) OID() GitObjectID { return r.oid }

func (r *GitCommitResolver) AbbreviatedOID() string {
	return string(r.oid)[:7]
}

func (r *GitCommitResolver) URL() string {
	url := r.repoResolver.url()
	url.Path += "/-/commit/" + r.inputRevOrImmutableRev()
	return url.String()
}

// inputRevOrImmutableRev returns the input revspec, if it is provided and nonempty. Otherwise it returns the
// canonical OID for the revision.
func (r *GitCommitResolver) inputRevOrImmutableRev() string {
	if r.inputRev != nil && *r.inputRev != "" {
		return *r.inputRev
	}
	return string(r.oid)
}
