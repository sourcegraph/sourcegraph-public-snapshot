package graphqlbackend

import (
	"context"
	"strings"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PerforceChangelistResolver struct {
	cid string
}

func toPerforceChangelistResolver(ctx context.Context, r *RepositoryResolver, commitBody string) (*PerforceChangelistResolver, error) {
	if source, err := r.SourceType(ctx); err != nil {
		return nil, err
	} else if *source != PerforceDepotSourceType {
		return nil, nil
	}

	changelistID, err := getP4ChangelistID(commitBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate perforceChangelistID")
	}

	return &PerforceChangelistResolver{cid: changelistID}, nil
}

func (r *PerforceChangelistResolver) CID() string {
	return r.cid
}

var p4FusionCommitSubjectPattern = lazyregexp.New(`^(\d+) - (.*)$`)

func parseP4FusionCommitSubject(subject string) (string, error) {
	matches := p4FusionCommitSubjectPattern.FindStringSubmatch(subject)
	if len(matches) != 3 {
		return "", errors.Newf("failed to parse commit subject %q for commit converted by p4-fusion", subject)
	}
	return matches[2], nil
}

// Either git-p4 or p4-fusion could be used to convert a perforce depot to a git repo. In which case the
// [git-p4: depot-paths = "//test-perms/": change = 83725]
// [p4-fusion: depot-paths = "//test-perms/": change = 80972]
var gitP4Pattern = lazyregexp.New(`\[(?:git-p4|p4-fusion): depot-paths = "(.*?)"\: change = (\d+)\]`)

func getP4ChangelistID(body string) (string, error) {
	matches := gitP4Pattern.FindStringSubmatch(body)
	if len(matches) != 3 {
		return "", errors.Newf("failed to retrieve changelist ID from commit body: %q", body)
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
