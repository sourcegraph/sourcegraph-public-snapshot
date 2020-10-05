package inference

import (
	"fmt"
	"testing"
)

func TestContainsSegment(t *testing.T) {
	testCases := []struct {
		path     string
		segment  string
		expected bool
	}{
		{path: "foo/bar/baz/bonk", segment: "foo", expected: true},
		{path: "foo/bar/baz/bonk", segment: "bar", expected: true},
		{path: "foo/bar/baz/bonk", segment: "baz", expected: true},
		{path: "foo/bar/baz/bonk", segment: "bonk", expected: true},
		{path: "foo/bar/baz/bonk", segment: "quux", expected: false},
		{path: "foo/bar/baz/bonk", segment: "honk", expected: false},
	}

	for _, testCase := range testCases {
		name := fmt.Sprintf("%s in %s", testCase.segment, testCase.path)

		t.Run(name, func(t *testing.T) {
			if value := containsSegment(testCase.path, testCase.segment); value != testCase.expected {
				t.Errorf("unexpected result. want=%v have=%v", testCase.expected, value)
			}
		})
	}
}
