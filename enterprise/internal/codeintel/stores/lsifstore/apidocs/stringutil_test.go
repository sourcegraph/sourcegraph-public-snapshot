package apidocs

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestTruncate(t *testing.T) {
	testCases := []struct {
		input      string
		limitBytes int
		want       autogold.Value
	}{
		{"", 5, autogold.Want("empty string", "")},
		{"1", 5, autogold.Want("tiny string not limited", "1")},
		{"1", 1, autogold.Want("tiny string limited", "1")},
		{"123456789", 5, autogold.Want("basic", "12345…")},
		{"👪👪", 5, autogold.Want("two four byte family characters, do not break them", "👪…")},
		{"1234👪👪56789", 5, autogold.Want("do not break unicode runes, exclude if they exceed byte limit", "1234…")},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := Truncate(tc.input, tc.limitBytes)
			tc.want.Equal(t, got)
		})
	}
}

func TestReverse(t *testing.T) {
	testCases := []struct {
		input string
		want  autogold.Value
	}{
		{"", autogold.Want("empty string", "")},
		{"h", autogold.Want("one character", "h")},
		{"asdf", autogold.Want("ascii", "fdsa")},
		{"as⃝df̅", autogold.Want("unicode with combining characters", "f̅ds⃝a")},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := Reverse(tc.input)
			tc.want.Equal(t, got)
		})
	}
}
