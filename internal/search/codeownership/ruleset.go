package codeownership

import (
	"bytes"
	"context"

	"github.com/hmarr/codeowners"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type Ruleset struct {
	codeownersRuleset codeowners.Ruleset
}
type Owners = []codeowners.Owner

func (r *Ruleset) Match(path string) (Owners, error) {
	if r.codeownersRuleset == nil {
		return Owners{}, nil
	}

	rule, err := r.codeownersRuleset.Match(path)
	if err != nil {
		return Owners{}, err
	}

	if rule == nil {
		return Owners{}, nil
	}

	// We directly return the codeowners.Owner struct to avoid creating
	// unnecessary copies. We also found that the longest list of owners
	// is less than 50.
	// c.f. https://github.com/sourcegraph/sourcegraph/pull/39250#discussion_r927942090
	return rule.Owners, nil
}

func NewRuleset(ctx context.Context, gitserver gitserver.Client, repoName api.RepoName, commitID api.CommitID) (Ruleset, error) {
	ruleset := Ruleset{}

	content, err := loadOwnershipFile(ctx, gitserver, repoName, commitID)
	if err != nil {
		return ruleset, err
	}
	if content == nil {
		return ruleset, nil
	}

	codeownersRuleset, err := codeowners.ParseFile(bytes.NewReader(content))
	if err != nil {
		return ruleset, err
	}

	ruleset.codeownersRuleset = codeownersRuleset

	return ruleset, nil
}

func loadOwnershipFile(ctx context.Context, gitserver gitserver.Client, repoName api.RepoName, commitID api.CommitID) ([]byte, error) {
	for _, path := range []string{"CODEOWNERS", ".github/CODEOWNERS", ".gitlab/CODEOWNERS", "docs/CODEOWNERS"} {
		content, err := gitserver.ReadFile(
			ctx,
			repoName,
			commitID,
			path,
			authz.DefaultSubRepoPermsChecker,
		)

		if err == nil && content != nil {
			return content, nil
		}
	}

	return nil, nil
}
