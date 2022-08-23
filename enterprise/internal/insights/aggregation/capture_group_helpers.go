package aggregation

// This logic is pulled from the compute package, with slight modifications.
// The intention is to not take a dependency on the compute package itself.

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Match struct {
	Value       string      `json:"value"`
	Range       Range       `json:"range"`
	Environment Environment `json:"environment"`
}

type Range struct {
	Start Location `json:"start"`
	End   Location `json:"end"`
}

type Environment map[string]Data

type Location struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Data struct {
	Value string `json:"value"`
	Range Range  `json:"range"`
}

func chunkContent(c result.ChunkMatch, r result.Range) string {
	// Set range relative to the start of the content.
	rr := r.Sub(c.ContentStart)
	return c.Content[rr.Start.Offset:rr.End.Offset]
}

func fromRegexpMatches(submatches []int, namedGroups []string, content string, range_ result.Range) Match {
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
		captureRange := newRange(range_.Start.Offset+start, range_.Start.Offset+end)

		if j == 0 {
			// The first submatch is the overall match
			// value. Don’t add this to the Environment
			firstValue = value
			firstRange = captureRange
			continue
		}

		var v string
		if namedGroups[j/2] == "" {
			v = strconv.Itoa(j / 2)
		} else {
			v = namedGroups[j/2]
		}
		env[v] = Data{Value: value, Range: captureRange}
	}
	return Match{Value: firstValue, Range: firstRange, Environment: env}
}

func newRange(startOffset, endOffset int) Range {
	return Range{
		Start: newLocation(-1, -1, startOffset),
		End:   newLocation(-1, -1, endOffset),
	}
}

func newLocation(line, column, offset int) Location {
	return Location{
		Offset: offset,
		Line:   line,
		Column: column,
	}
}

func toTextResult(content string, matchPattern MatchPattern, outputPattern, separator, selector string) (string, error) {
	if selector != "" {
		// Don't run the search pattern over the search result content
		// when there's an explicit `select:` value.
		return outputPattern, nil
	}

	return output(content, matchPattern, outputPattern, separator)
}

func output(fragment string, matchPattern MatchPattern, replacePattern string, separator string) (string, error) {
	var newContent string
	switch match := matchPattern.(type) {
	case *Regexp:
		newContent = substituteRegexp(fragment, match.Value, replacePattern, separator)
	}

	return newContent, nil
}

func substituteRegexp(content string, match *regexp.Regexp, replacePattern, separator string) string {
	var b strings.Builder
	for _, submatches := range match.FindAllStringSubmatchIndex(content, -1) {
		b.Write(match.ExpandString([]byte{}, replacePattern, content, submatches))
		b.WriteString(separator)
	}
	return b.String()
}

type MatchPattern interface {
	pattern()
	String() string
}

func (Regexp) pattern() {}
func (Comby) pattern()  {}

type Regexp struct {
	Value *regexp.Regexp
}

type Comby struct {
	Value string
}

func (p Regexp) String() string {
	return p.Value.String()
}

func (p Comby) String() string {
	return p.Value
}

func toRegexpPattern(value string) (MatchPattern, error) {
	rp, err := regexp.Compile(value)
	if err != nil {
		return nil, errors.Wrap(err, "compute endpoint")
	}
	return &Regexp{Value: rp}, nil
}
