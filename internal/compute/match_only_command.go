package compute

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type MatchOnly struct {
	MatchPattern MatchPattern
}

func (c *MatchOnly) String() string {
	return fmt.Sprintf("Match only: %s", c.MatchPattern.String())
}

func fromRegexpMatches(matches [][]int, namedGroups []string, lineValue string, lineNumber int) Match {
	env := make(Environment)
	var firstValue string
	var firstRange Range
	for _, m := range matches {
		// iterate over pairs of offsets. Cf. FindAllStringSubmatchIndex
		// https://pkg.go.dev/regexp#Regexp.FindAllStringSubmatchIndex.
		for j := 0; j < len(m); j += 2 {
			start := m[j]
			end := m[j+1]
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
	}
	return Match{Value: firstValue, Range: firstRange, Environment: env}
}

func FromFileMatch(fm *result.FileMatch, r *regexp.Regexp) *MatchContext {
	matches := make([]Match, 0, len(fm.LineMatches))
	for _, l := range fm.LineMatches {
		regexpMatches := r.FindAllStringSubmatchIndex(l.Preview, -1)
		if len(regexpMatches) > 0 {
			matches = append(matches, fromRegexpMatches(regexpMatches, r.SubexpNames(), l.Preview, int(l.LineNumber)))
		}
	}
	return &MatchContext{Matches: matches, Path: fm.Path}
}
