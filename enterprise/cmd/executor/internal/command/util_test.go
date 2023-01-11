package command

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFlatten(t *testing.T) {
	actual := flatten(
		"foo",
		[]string{"bar", "baz"},
		[]string{"bonk", "quux"},
	)

	expected := []string{
		"foo",
		"bar", "baz",
		"bonk", "quux",
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected slice (-want +got):\n%s", diff)
	}
}

func TestIntersperse(t *testing.T) {
	actual := intersperse("-e", []string{
		"A=B",
		"C=D",
		"E=F",
	})

	expected := []string{
		"-e", "A=B",
		"-e", "C=D",
		"-e", "E=F",
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected slice (-want +got):\n%s", diff)
	}
}

func TestQuoteEnv(t *testing.T) {
	tests := []struct {
		in   []string
		want []string
	}{
		{
			in:   []string{"FOO=bar"},
			want: []string{"FOO=bar"},
		},
		{
			in:   []string{"FOO=bar foo bar"},
			want: []string{`FOO='bar foo bar'`},
		},

		{
			in:   []string{"HOME=computer", "FOO=bar"},
			want: []string{"HOME=computer", "FOO=bar"},
		},
		{
			in:   []string{"HOME=compute r", "FOO=bar foo bar"},
			want: []string{`HOME='compute r'`, `FOO='bar foo bar'`},
		},
		{
			in:   []string{"FOO=bar -e 31337=H4XX0R"},
			want: []string{`FOO='bar -e 31337=H4XX0R'`},
		},
		{
			in:   []string{`FOO=bar -e "shell-h4xx0r"`},
			want: []string{`FOO='bar -e "shell-h4xx0r"'`},
		},
	}

	for _, tt := range tests {
		got := quoteEnv(tt.in)

		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("unexpected slice (-want +got):\n%s", diff)
		}
	}
}
