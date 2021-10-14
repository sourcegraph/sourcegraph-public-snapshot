package lsifstore

import (
	"testing"

	"github.com/hexops/autogold"
)

func Test_lexemes(t *testing.T) {
	testCases := []struct {
		input string
		want  autogold.Value
	}{
		{"", autogold.Want("empty string", []string{})},
		{"f", autogold.Want("single alphabetical", []string{"f"})},
		{".", autogold.Want("single punctuation", []string{"."})},
		{"f.", autogold.Want("single alphabetical and punctuation", []string{"f", "."})},
		{"foo.bar.baz", autogold.Want("basic", []string{"foo", ".", "bar", ".", "baz"})},
		{"foo::bar'a new Baz().bar//efg", autogold.Want("complex", []string{
			"foo", ":", ":", "bar", "'", "a", "new", "Baz", "(",
			")",
			".",
			"bar",
			"/",
			"/",
			"efg",
		})},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := lexemes(tc.input)
			tc.want.Equal(t, got)
		})
	}
}

func Test_reverse(t *testing.T) {
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
			got := reverse(tc.input)
			tc.want.Equal(t, got)
		})
	}
}

func Test_textSearchVector(t *testing.T) {
	testCases := []struct {
		input string
		want  autogold.Value
	}{
		{"", autogold.Want("empty string", "")},
		{"hello world", autogold.Want("english", "hello:1 world:2")},
		{"http.Router", autogold.Want("basic", "http:1 .:2 Router:3")},
		{"go github.com/golang/go private struct http.Router", autogold.Want("complex", "go:1 github:2 .:3 com:4 /:5 golang:6 /:7 go:8 private:9 struct:10 http:11 .:12 Router:13")},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := textSearchVector(tc.input)
			tc.want.Equal(t, got)
		})
	}
}
