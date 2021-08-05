package compute

import (
	"fmt"
	"regexp"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

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

type Result struct {
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

func ofRegexpMatches(matches [][]int, lineValue string, lineNumber int) Match {
	env := make(Environment)
	var firstValue string
	var firstRange Range
	for _, m := range matches {
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
			v := fmt.Sprintf("$%d", j/2)
			env[v] = Data{Value: value, Range: range_}
		}
	}
	return Match{Value: firstValue, Range: firstRange, Environment: env}
}

func ofFileMatches(fm *result.FileMatch, r *regexp.Regexp) *Result {
	var matches []Match
	for _, l := range fm.LineMatches {
		regexpMatches := r.FindAllStringSubmatchIndex(l.Preview, -1)
		matches = append(matches, ofRegexpMatches(regexpMatches, l.Preview, int(l.LineNumber)))
	}
	return &Result{Matches: matches, Path: fm.Path}
}
