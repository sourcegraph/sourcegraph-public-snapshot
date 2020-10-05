package inference

import (
	"regexp"
	"strings"
)

var Patterns = func() *regexp.Regexp {
	var patterns []string
	for _, recognizer := range Recognizers {
		patterns = append(patterns, recognizer.Patterns()...)
	}

	return regexp.MustCompile(strings.Join(patterns, "|"))
}()

// TODO - convenience functions too
