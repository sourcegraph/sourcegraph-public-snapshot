package inference

import (
	"regexp"
	"strings"
)

// Patterns is a regular expression that matches any input that would also match
// any registered recognizer pattern.
var Patterns = func() *regexp.Regexp {
	var patterns []string
	for _, recognizer := range Recognizers {
		for _, pattern := range recognizer.Patterns() {
			patterns = append(patterns, pattern.String())
		}
	}

	return regexp.MustCompile(strings.Join(patterns, "|"))
}()

// rawPattern creates a regular expression matching the given string literal.
func rawPattern(s string) *regexp.Regexp {
	return regexp.MustCompile(regexp.QuoteMeta(s))
}

// prefixPattern creates a regular expression that matches strings beginning with
// the given pattern.
func prefixPattern(pattern *regexp.Regexp) *regexp.Regexp {
	return regexp.MustCompile("^" + pattern.String())
}

// suffixPattern creates a regular expression that matches strings ending with the
// given pattern.
func suffixPattern(pattern *regexp.Regexp) *regexp.Regexp {
	return regexp.MustCompile(pattern.String() + "$")
}

// extensionPattern creates a regular expression that matches paths with the given
// extension. The extension separator is added automatically.
func extensionPattern(pattern *regexp.Regexp) *regexp.Regexp {
	return suffixPattern(regexp.MustCompile("(^|/)[^/]+." + pattern.String()))
}

// pathPattern creates a regular expression that matches paths with the given basename.
func pathPattern(pattern *regexp.Regexp) *regexp.Regexp {
	return suffixPattern(regexp.MustCompile("(^|/)" + pattern.String()))
}
