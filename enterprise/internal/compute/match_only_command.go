package compute

import (
	"context"
	"fmt"
	"strconv"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
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

func fromRegexpMatches(submatches []int, namedGroups []string, lineValue string, lineNumber int) Match {
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
		value := lineValue[start:end]
		range_ := newRange(lineNumber, lineNumber, start, end)

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

func matchOnly(fm *result.FileMatch, r *regexp.Regexp) *MatchContext {
	lineMatches := fm.ChunkMatches.AsLineMatches()
	matches := make([]Match, 0, len(lineMatches))
	for _, l := range lineMatches {
		for _, submatches := range r.FindAllStringSubmatchIndex(l.Preview, -1) {
			matches = append(matches, fromRegexpMatches(submatches, r.SubexpNames(), l.Preview, int(l.LineNumber)))
		}
	}
	return &MatchContext{Matches: matches, Path: fm.Path, RepositoryID: int32(fm.Repo.ID), Repository: string(fm.Repo.Name)}
}

func (c *MatchOnly) Run(_ context.Context, db database.DB, r result.Match) (Result, error) {
	switch m := r.(type) {
	case *result.FileMatch:
		return matchOnly(m, c.ComputePattern.(*Regexp).Value), nil
	}
	return nil, nil
}
