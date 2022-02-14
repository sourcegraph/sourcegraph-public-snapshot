package inference

import (
	"regexp"
	"strings"
)

// Patterns is a regular expression that matches any input that would also match
// any registered recognizer pattern.
var Patterns = func() *regexp.Regexp {
	var patterns []*regexp.Regexp
	for _, recognizer := range Recognizers {
		patterns = append(patterns, recognizer.Patterns()...)
	}
	return OrPattern(patterns)
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

// OrPattern [r1, r2, r3, ...] is equivalent to r1|r2|r3|... .
func OrPattern(patterns []*regexp.Regexp) *regexp.Regexp {
	var patternStrings []string
	for _, pattern := range patterns {
		patternStrings = append(patternStrings, pattern.String())
	}
	return regexp.MustCompile(strings.Join(patternStrings, "|"))
}
