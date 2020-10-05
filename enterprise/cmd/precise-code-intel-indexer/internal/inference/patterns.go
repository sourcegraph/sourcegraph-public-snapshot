package inference

import (
	"regexp"
	"strings"
)

// Patterns is a regular expression that matches any input that would also match
// any registered recognizer pattern.``
var Patterns = func() *regexp.Regexp {
	var patterns []string
	for _, recognizer := range Recognizers {
		patterns = append(patterns, recognizer.Patterns()...)
	}

	return regexp.MustCompile(strings.Join(patterns, "|"))
}()

// TODO - convenience functions too
