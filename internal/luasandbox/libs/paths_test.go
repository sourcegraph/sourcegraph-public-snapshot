package libs

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDirWithoutDot(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"file.txt", ""},
		{"simple/file.txt", "simple"},
		{"longer/ancestor/paths/file.txt", "longer/ancestor/paths"},
		{"longer/ancestor/paths/subdir", "longer/ancestor/paths"},
	}

	t.Run("edge cases", func(t *testing.T) {
		for _, input := range []string{"", "/", "."} {
			if actual := dirWithoutDot(input); actual != "" {
				t.Errorf("unexpected dirname: want=%s got=%s", "", actual)
			}
		}
	})

	t.Run("un-rooted", func(t *testing.T) {
		for _, testCase := range testCases {
			if actual := dirWithoutDot(testCase.input); actual != testCase.expected {
				t.Errorf("unexpected dirname: want=%s got=%s", testCase.expected, actual)
			}
		}
	})

	t.Run("rooted", func(t *testing.T) {
		for _, testCase := range testCases {
			if actual := dirWithoutDot("/" + testCase.input); actual != testCase.expected {
				t.Errorf("unexpected dirname: want=%s got=%s", testCase.expected, actual)
			}
		}
	})
}

func TestAncestorDirs(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{"", []string{""}},
		{"file.txt", []string{""}},
		{"simple/file.txt", []string{"simple", ""}},
		{"longer/ancestor/paths/file.txt", []string{"longer/ancestor/paths", "longer/ancestor", "longer", ""}},
	}

	t.Run("un-rooted", func(t *testing.T) {
		for _, testCase := range testCases {
			if diff := cmp.Diff(testCase.expected, ancestorDirs(testCase.input)); diff != "" {
				t.Errorf("unexpected ancestor for %q dirs (-want +got):\n%s", testCase.input, diff)
			}
		}
	})

	t.Run("rooted", func(t *testing.T) {
		for _, testCase := range testCases {
			if diff := cmp.Diff(testCase.expected, ancestorDirs("/"+testCase.input)); diff != "" {
				t.Errorf("unexpected ancestor for %q dirs (-want +got):\n%s", testCase.input, diff)
			}
		}
	})
}
