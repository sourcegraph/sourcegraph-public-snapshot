package compute

import (
	"fmt"
	"regexp"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

// Location represents the position in a text file, which may be an absolute
// offset or line/column pair. Offsets can be converted to line/columns or vice
// versa when the input file is available. We represent the possibility, but not
// the requirement, of representing either offset or line/column in this data
// type because tools or processes may expose only, e.g., offsets for
// performance reasons (e.g., parsing) and leave conversion (which has
// performance implications) up to the client. Nevertheless, from a usability
// perspective, it is advantageous to represent both possibilities in a single
// type. Conventionally, "null" values may be represented with -1.
type Location struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Range struct {
	Start Location `json:"start"`
	End   Location `json:"end"`
}

type Data struct {
	Value string `json:"value"`
	Range Range  `json:"range"`
}

type Environment map[string]Data

type Match struct {
	Value       string      `json:"value"`
	Range       Range       `json:"range"`
	Environment Environment `json:"environment"`
}

type MatchContext struct {
	Matches []Match `json:"matches"`
	Path    string  `json:"path"`
}

func newLocation(line, column, offset int) Location {
	return Location{
		Offset: offset,
		Line:   line,
		Column: column,
	}
}

func newRange(startLine, endLine, startColumn, endColumn int) Range {
	return Range{
		Start: newLocation(startLine, startColumn, -1),
		End:   newLocation(endLine, endColumn, -1),
	}
}

func ofRegexpMatches(matches [][]int, namedGroups []string, lineValue string, lineNumber int) Match {
	env := make(Environment)
	var firstValue string
	var firstRange Range
	for _, m := range matches {
		// iterate over pairs of offsets. Cf. FindAllStringSubmatchIndex
		// https://pkg.go.dev/regexp#Regexp.FindAllStringSubmatchIndex.
		for j := 0; j < len(m); j += 2 {
			start := m[j]
			end := m[j+1]
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
				v = fmt.Sprintf("$%d", j/2)
			} else {
				v = namedGroups[j/2]
			}
			env[v] = Data{Value: value, Range: range_}
		}
	}
	return Match{Value: firstValue, Range: firstRange, Environment: env}
}

func ofFileMatches(fm *result.FileMatch, r *regexp.Regexp) *MatchContext {
	matches := make([]Match, 0, len(fm.LineMatches))
	for _, l := range fm.LineMatches {
		regexpMatches := r.FindAllStringSubmatchIndex(l.Preview, -1)
		if len(regexpMatches) > 0 {
			matches = append(matches, ofRegexpMatches(regexpMatches, r.SubexpNames(), l.Preview, int(l.LineNumber)))
		}
	}
	return &MatchContext{Matches: matches, Path: fm.Path}
}
