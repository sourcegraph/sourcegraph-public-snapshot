package codeownership

import (
	"context"
	"strings"
	"time"

	"github.com/hmarr/codeowners"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type Ruleset struct {
	codeownersRuleset codeowners.Ruleset
}
type Owners []string

func (r *Ruleset) Match(path string) (Owners, error) {
	if r.codeownersRuleset == nil {
		return []string{}, nil
	}

	rule, err := r.codeownersRuleset.Match(path)
	if err != nil {
		return []string{}, err
	}

	if rule == nil {
		return []string{}, nil
	}

	owners := make(Owners, len(rule.Owners))
	for i, owner := range rule.Owners {
		owners[i] = owner.String()
	}
	return owners, nil
}

func NewRuleset(db database.DB, repoName api.RepoName, commitID api.CommitID) (Ruleset, error) {
	var content []byte
	var err error

	ruleset := Ruleset{}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	content, err = loadOwnershipFile(ctx, db, repoName, commitID)
	if err != nil {
		return ruleset, err
	}
	if content == nil {
		return ruleset, nil
	}

	codeownersRuleset, err := codeowners.ParseFile(strings.NewReader(string(content)))
	if err != nil {
		return ruleset, err
	}

	ruleset.codeownersRuleset = codeownersRuleset

	return ruleset, nil
}

func loadOwnershipFile(ctx context.Context, db database.DB, repoName api.RepoName, commitID api.CommitID) ([]byte, error) {
	for _, path := range []string{"CODEOWNERS", ".github/CODEOWNERS", ".gitlab/CODEOWNERS", "docs/CODEOWNERS"} {
		content, err := gitserver.NewClient(db).ReadFile(
			ctx,
			repoName,
			commitID,
			path,
			authz.DefaultSubRepoPermsChecker,
		)

		if err != nil {
			return nil, err
		}

		if content != nil {
			return content, nil
		}
	}

	return nil, nil
}
