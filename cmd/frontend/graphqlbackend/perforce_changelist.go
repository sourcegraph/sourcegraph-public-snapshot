package graphqlbackend

import (
	"context"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PerforceChangelistResolver struct {
	// logger is a logger - what more needs to be said. 🪵
	logger log.Logger

	// repositoryResolver is the backlink to which this change list belongs.
	repositoryResolver *RepositoryResolver

	// cid is the changelist ID.
	cid string
	// canonicalURL is the canonical URL of this changelist ID, similar to the canonical URL of a Git commit.
	canonicalURL string
	// commitSHA is the corresponding commit SHA. This is required to look up the commitID object using ResolveRev.
	commitSHA string

	// commitID is set when the Commit property is accessed on the resolver.
	commitID api.CommitID
	// commitOnce will ensure that we resolve the revision only once.
	commitOnce sync.Once
	// commitErr is used to return an error that may have occured during resolving the revision when
	// the Commit property is looked up on the resolver.
	commitErr error
}

func newPerforceChangelistResolver(r *RepositoryResolver, changelistID, commitSHA string) *PerforceChangelistResolver {
	repoURL := r.url()

	// Example: /perforce.sgdev.org/foobar/-/changelist/99999
	canonicalURL := filepath.Join(repoURL.Path, "-", "changelist", changelistID)

	return &PerforceChangelistResolver{
		logger:             r.logger.Scoped("PerforceChangelistResolver"),
		repositoryResolver: r,
		cid:                changelistID,
		commitSHA:          commitSHA,
		canonicalURL:       canonicalURL,
	}
}

func toPerforceChangelistResolver(ctx context.Context, gcr *GitCommitResolver) (*PerforceChangelistResolver, error) {
	if source, err := gcr.repoResolver.SourceType(ctx); err != nil {
		return nil, err
	} else if *source != PerforceDepotSourceType {
		return nil, nil
	}

	commit, err := gcr.resolveCommit(ctx)
	if err != nil {
		return nil, err
	}

	changelistID, err := perforce.GetP4ChangelistID(commit.Message.Body())
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate perforceChangelistID")
	}

	return newPerforceChangelistResolver(gcr.repoResolver, changelistID, string(commit.ID)), nil
}

func (r *PerforceChangelistResolver) CID() string {
	return r.cid
}

func (r *PerforceChangelistResolver) CanonicalURL() string {
	return r.canonicalURL
}

func (r *PerforceChangelistResolver) cidURL() *url.URL {
	repoURL := r.repositoryResolver.url()
	// We don't expect cid to be empty, but guard against any potential bugs.
	if r.cid != "" {
		repoURL.Path += "@" + r.cid
	}
	return repoURL
}

func (r *PerforceChangelistResolver) Commit(ctx context.Context) (_ *GitCommitResolver, err error) {
	repoResolver := r.repositoryResolver
	r.commitOnce.Do(func() {
		r.commitID, r.commitErr = backend.NewRepos(
			r.logger,
			repoResolver.db,
			repoResolver.gitserverClient,
		).ResolveRev(ctx, repoResolver.name, r.commitSHA)
	})

	if r.commitErr != nil {
		return nil, r.commitErr
	}

	commitResolver := NewGitCommitResolver(repoResolver.db, repoResolver.gitserverClient, r.repositoryResolver, r.commitID, nil)
	commitResolver.inputRev = &r.commitSHA
	return commitResolver, nil
}

var p4FusionCommitSubjectPattern = lazyregexp.New(`^(\d+) - (.*)$`)

func parseP4FusionCommitSubject(subject string) (string, error) {
	matches := p4FusionCommitSubjectPattern.FindStringSubmatch(subject)
	if len(matches) != 3 {
		return "", errors.Newf("failed to parse commit subject %q for commit converted by p4-fusion", subject)
	}
	return matches[2], nil
}

// maybeTransformP4Body is used for special handling of perforce depots converted to git using
// p4-fusion or git-p4. We want to strip out the generated commit message and use the original
// We handle both p4-fusion and git-p4 so that we stripe the system message from both.
func maybeTransformP4Body(body string) *string {
	if idx := strings.Index(body, "[p4-fusion"); idx != -1 {
		body = body[:idx]
	} else if idx := strings.Index(body, "[git-p4"); idx != -1 {
		body = body[:idx]
	}
	trimmedBody := strings.TrimSpace(body)
	return &trimmedBody

}

// maybeTransformP4Subject is used for special handling of perforce depots converted to git using
// p4-fusion. We want to parse and use the subject from the original changelist and not the subject
// that is generated during the conversion.
//
// For depots converted with git-p4, this special handling is NOT required.
func maybeTransformP4Subject(ctx context.Context, repoResolver *RepositoryResolver, commit *gitdomain.Commit) *string {
	if repoResolver.isPerforceDepot(ctx) && strings.Contains(commit.Message.Body(), "[p4-fusion") {
		subject, err := parseP4FusionCommitSubject(commit.Message.Subject())
		if err == nil {
			return &subject
		} else {
			// If parsing this commit message fails for some reason, log the reason and fall-through
			// to return the the original git-commit's subject instead of a hard failure or an empty
			// subject.
			repoResolver.logger.Error("failed to parse p4 fusion commit subject", log.Error(err))
		}
	}

	return nil
}
