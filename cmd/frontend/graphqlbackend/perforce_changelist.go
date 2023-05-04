package graphqlbackend

import (
	"context"

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
