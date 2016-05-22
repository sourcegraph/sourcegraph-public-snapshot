package textutil

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestFirstSentence(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello. there.", "hello."},
		{"", ""},
		{"a...b c. ok.", "a...b c."},
		{"ok a.b.c.", "ok a.b.c."},
		{"ok.\nfoo.", "ok."},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa..."},
		{"aaaaaa", "aaaaaa"},
		{".", "."},
	}
	for _, test := range tests {
		got := FirstSentence(test.input)
		if test.want != got {
			t.Errorf("%q: want %q, got %q", test.input, test.want, got)
		}
	}
}

func TestFirstName(t *testing.T) {
	tests := map[string]string{
		"Foo Bar": "Foo",
		"Foo":     "Foo",
		"":        "",
		" ":       "",
	}
	for in, expOut := range tests {
		if out := FirstName(&sourcegraph.User{Name: in}); out != expOut {
			t.Errorf("expected %s, got %s", expOut, out)
		}
	}
}

func TestFirstNameOrLogin(t *testing.T) {
	tests := []struct {
		in  *sourcegraph.User
		out string
	}{
		{in: &sourcegraph.User{Name: "foo"}, out: "foo"},
		{in: &sourcegraph.User{Name: "", Login: "bar"}, out: "bar"},
	}
	for _, test := range tests {
		if out := FirstNameOrLogin(test.in); out != test.out {
			t.Errorf("expected %s, but got %s", test.out, out)
		}
	}
}

func TestShortCommitMessage(t *testing.T) {
	tests := []struct {
		in, out string
		n       int
	}{
		{in: "under-80\nbut\nmultiple\nlines", out: "under-80", n: 80},
		{
			in:  "this text is exactly 29 runes and from there the sentence just goes downhill\n but that's okay because it gets stripped off.",
			out: "this text is exactly 29 rune…",
			n:   29,
		},
	}
	for _, test := range tests {
		if out := ShortCommitMessage(test.n, test.in); out != test.out {
			t.Errorf("expected %q, but got %q", test.out, out)
		}
	}
}
