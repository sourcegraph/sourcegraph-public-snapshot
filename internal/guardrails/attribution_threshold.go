package guardrails

import "strings"

type SnippetLowerBound struct {
	linesLowerBound int
}

func NewThreshold() SnippetLowerBound {
	return SnippetLowerBound{
		linesLowerBound: 10,
	}
}

// ShouldSearch discerns whether attribution search should run
// for the given snippet at all, or is it too small :(
func (t SnippetLowerBound) ShouldSearch(snippet string) bool {
	// Nine breaklines is ten lines, so offset lines cound by one.
	return strings.Count(snippet, "\n") >= (t.LinesLowerBound() - 1)
}

// LinesLowerBound is the minimum number of lines that are considered
// for attribution search. 0 and 1 means no limit.
func (t SnippetLowerBound) LinesLowerBound() int {
	return t.linesLowerBound
}
