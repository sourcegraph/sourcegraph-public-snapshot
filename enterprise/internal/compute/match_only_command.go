package compute

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type MatchOnly struct {
	SearchPattern MatchPattern

	// ComputePattern is the valid, semantically-equivalent representation
	// of MatchPattern that mirrors implicit Sourcegraph search behavior
	// (e.g., default case insensitivity), but which may differ
	// syntactically (e.g., by wrapping a pattern in (?i:<MatchPattern>).
	ComputePattern MatchPattern
}

func (c *MatchOnly) ToSearchPattern() string {
	return c.SearchPattern.String()
}

func (c *MatchOnly) String() string {
	return fmt.Sprintf(
		"Match only search pattern: %s, compute pattern: %s",
		c.SearchPattern.String(),
		c.ComputePattern.String(),
	)
}

func fromRegexpMatches(submatches []int, namedGroups []string, content string) Match {
	env := make(Environment)
	var firstValue string
	var firstRange Range
	// iterate over pairs of offsets. Cf. FindAllStringSubmatchIndex
	// https://pkg.go.dev/regexp#Regexp.FindAllStringSubmatchIndex.
	for j := 0; j < len(submatches); j += 2 {
		start := submatches[j]
		end := submatches[j+1]
		if start == -1 || end == -1 {
			// The entire regexp matched, but a capture
			// group inside it did not. Ignore this entry.
			continue
		}
		value := content[start:end]
		// NOTE: Since we match potentially multiline fragments, we
		// cannot assume that we are matching only on one line. The
		// regex library doesn't do line accounting and doesn't give
		// this information about matches. So, we cannot get the line
		// range/column range without doing real heavy work. We
		// currently don't do that heavy work and only expose the range
		// offsets. https://github.com/sourcegraph/sourcegraph/issues/30711
		range_ := newOffsetRange(start, end)

		if j == 0 {
			// The first submatch is the overall match
			// value. Don't add this to the Environment
			firstValue = value
			firstRange = range_
			continue
		}

		var v string
		if namedGroups[j/2] == "" {
			v = strconv.Itoa(j / 2)
		} else {
			v = namedGroups[j/2]
		}
		env[v] = Data{Value: value, Range: range_}
	}
	return Match{Value: firstValue, Range: firstRange, Environment: env}
}

func matchOnly(ctx context.Context, fm *result.FileMatch, r *regexp.Regexp) (*MatchContext, error) {
	content, err := git.ReadFile(ctx, fm.Repo.Name, fm.CommitID, fm.Path, 0, authz.DefaultSubRepoPermsChecker)
	if err != nil {
		return nil, err
	}

	matches := []Match{}
	for _, submatches := range r.FindAllStringSubmatchIndex(string(content), -1) {
		matches = append(matches, fromRegexpMatches(submatches, r.SubexpNames(), string(content)))
	}
	return &MatchContext{Matches: matches, Path: fm.Path}, nil
}

func (c *MatchOnly) Run(ctx context.Context, r result.Match) (Result, error) {
	switch m := r.(type) {
	case *result.FileMatch:
		return matchOnly(ctx, m, c.ComputePattern.(*Regexp).Value)
	}
	return nil, nil
}
