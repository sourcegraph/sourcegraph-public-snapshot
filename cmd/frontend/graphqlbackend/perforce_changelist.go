package graphqlbackend

import (
	"context"
	"strings"
	"sync"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PerforceChangelistResolver struct {
	logger log.Logger

	repositoryResolver *RepositoryResolver

	cid          string
	canonicalURL string
	commitSHA    string

	commitID   api.CommitID
	commitOnce sync.Once
	commitErr  error
}

func newPerforceChangelistResolver(ctx context.Context, r *RepositoryResolver, changelistID, commitSHA string) *PerforceChangelistResolver {
	repoURL := r.url()
	canonicalURL := repoURL.Path + "/-/changelist/" + changelistID

	return &PerforceChangelistResolver{
		logger:             r.logger.Scoped("PerforceChangelistResolver", "resolve a specific changelist"),
		repositoryResolver: r,
		cid:                changelistID,
		commitSHA:          commitSHA,
		canonicalURL:       canonicalURL,
	}
}

func toPerforceChangelistResolver(ctx context.Context, r *RepositoryResolver, commit *gitdomain.Commit) (*PerforceChangelistResolver, error) {
	if source, err := r.SourceType(ctx); err != nil {
		return nil, err
	} else if *source != PerforceDepotSourceType {
		return nil, nil
	}

	changelistID, err := perforce.GetP4ChangelistID(commit.Message.Body())
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate perforceChangelistID")
	}

	return newPerforceChangelistResolver(ctx, r, changelistID, string(commit.ID)), nil
}

func (r *PerforceChangelistResolver) CID() string {
	return r.cid
}

func (r *PerforceChangelistResolver) CanonicalURL() string {
	return r.canonicalURL
}

func (r *PerforceChangelistResolver) Commit(ctx context.Context) (_ *GitCommitResolver, err error) {
	repoResolver := r.repositoryResolver
	r.commitOnce.Do(func() {
		repo, err := repoResolver.repo(ctx)
		if err != nil {
			r.commitErr = err
			return
		}

		r.commitID, r.commitErr = backend.NewRepos(
			r.logger,
			repoResolver.db,
			repoResolver.gitserverClient,
		).ResolveRev(ctx, repo, r.commitSHA)
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

// maybeTransformP4Subject is used for special handling of perforce depots converted to git using
// p4-fusion. We want to parse and use the subject from the original changelist and not the subject
// that is generated during the conversion.
//
// For depots converted with git-p4, this special handling is NOT required.
func maybeTransformP4Subject(ctx context.Context, repoResolver *RepositoryResolver, commit *gitdomain.Commit) *string {
	if repoResolver.isPerforceDepot(ctx) && strings.HasPrefix(commit.Message.Body(), "[p4-fusion") {
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
