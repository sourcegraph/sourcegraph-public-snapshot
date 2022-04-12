package inference

import (
	"fmt"
	"strings"
	"testing"

	"github.com/grafana/regexp"
)

type PathTestCase = struct {
	path     string
	expected bool
}

// Test if all test cases match against any pattern in patterns.
func testLangPatterns(t *testing.T, patterns []*regexp.Regexp, testCases []PathTestCase) {
	var patternStrings []string
	for _, pattern := range patterns {
		patternStrings = append(patternStrings, pattern.String())
	}
	var pattern = regexp.MustCompile(strings.Join(patternStrings, "|"))
	for _, testCase := range testCases {
		if pattern.MatchString(testCase.path) {
			if !testCase.expected {
				t.Error(fmt.Sprintf("did not expect match: %s", testCase.path))
			}
		} else if testCase.expected {
			t.Error(fmt.Sprintf("expected match: %s", testCase.path))
		}
	}
}
