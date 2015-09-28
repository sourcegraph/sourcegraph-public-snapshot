package annotate

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"text/template"
	"unicode/utf8"
)

var saveExp = flag.Bool("exp", false, "overwrite all expected output files with actual output (returning a failure)")
var match = flag.String("m", "", "only run tests whose name contains this string")

func TestAnnotate(t *testing.T) {
	tests := map[string]struct {
		input   string
		anns    Annotations
		want    string
		wantErr error
	}{
		"empty and unannotated": {"", nil, "", nil},
		"unannotated":           {"a⌘b", nil, "a⌘b", nil},

		// The docs say "Annotating an empty byte array always returns an empty
		// byte array.", which is arbitrary but makes implementation easier.
		"empty annotated": {"", Annotations{{0, 0, []byte("["), []byte("]"), 0}}, "", nil},

		"zero-length annotations": {
			"aaaa",
			Annotations{
				{0, 0, []byte("<b>"), []byte("</b>"), 0},
				{0, 0, []byte("<i>"), []byte("</i>"), 0},
				{2, 2, []byte("<i>"), []byte("</i>"), 0},
			},
			"<b></b><i></i>aa<i></i>aa",
			nil,
		},
		"1 annotation": {"a", Annotations{{0, 1, []byte("["), []byte("]"), 0}}, "[a]", nil},
		"nested": {
			"abc",
			Annotations{
				{0, 3, []byte("["), []byte("]"), 0},
				{1, 2, []byte("<"), []byte(">"), 0},
			},
			"[a<b>c]",
			nil,
		},
		"nested 1": {
			"abcd",
			Annotations{
				{0, 4, []byte("<1>"), []byte("</1>"), 0},
				{1, 3, []byte("<2>"), []byte("</2>"), 0},
				{2, 2, []byte("<3>"), []byte("</3>"), 0},
			},
			"<1>a<2>b<3></3>c</2>d</1>",
			nil,
		},
		"same range": {
			"ab",
			Annotations{
				{0, 2, []byte("["), []byte("]"), 0},
				{0, 2, []byte("<"), []byte(">"), 0},
			},
			"[<ab>]",
			nil,
		},
		"same range (with WantInner)": {
			"ab",
			Annotations{
				{0, 2, []byte("["), []byte("]"), 1},
				{0, 2, []byte("<"), []byte(">"), 0},
			},
			"<[ab]>",
			nil,
		},
		"unicode content": {
			"abcdef⌘vwxyz",
			Annotations{
				{6, 9, []byte("<a>"), []byte("</a>"), 0},
				{10, 12, []byte("<b>"), []byte("</b>"), 0},
				{0, 13, []byte("<c>"), []byte("</c>"), 0},
			},
			"<c>abcdef<a>⌘</a>v<b>wx</b>y</c>z",
			nil,
		},
		"remainder": {
			"xyz",
			Annotations{
				{0, 2, []byte("<b>"), []byte("</b>"), 0},
				{0, 1, []byte("<c>"), []byte("</c>"), 0},
			},
			"<b><c>x</c>y</b>z",
			nil,
		},

		// Errors
		"start oob": {"a", Annotations{{-1, 1, []byte("<"), []byte(">"), 0}}, "<a>", ErrStartOutOfBounds},
		"start oob (multiple)": {
			"a",
			Annotations{
				{-3, 1, []byte("1"), []byte(""), 0},
				{-3, 1, []byte("2"), []byte(""), 0},
				{-1, 1, []byte("3"), []byte(""), 0},
			},
			"123a",
			ErrStartOutOfBounds,
		},
		"end oob": {"a", Annotations{{0, 3, []byte("<"), []byte(">"), 0}}, "<a>", ErrEndOutOfBounds},
		"end oob (multiple)": {
			"ab",
			Annotations{
				{0, 3, []byte(""), []byte("1"), 0},
				{1, 3, []byte(""), []byte("2"), 0},
				{0, 5, []byte(""), []byte("3"), 0},
			},
			"ab213",
			ErrEndOutOfBounds,
		},
	}
	for label, test := range tests {
		if *match != "" && !strings.Contains(label, *match) {
			continue
		}

		sort.Sort(Annotations(test.anns))

		got, err := Annotate([]byte(test.input), test.anns, nil)
		if err != test.wantErr {
			if test.wantErr == nil {
				t.Errorf("%s: Annotate: %s", label, err)
			} else {
				t.Errorf("%s: Annotate: got error %v, want %v", label, err, test.wantErr)
			}
			continue
		}
		if string(got) != test.want {
			t.Errorf("%s: Annotate: got %q, want %q", label, got, test.want)
			continue
		}
	}
}

func TestAnnotate_Files(t *testing.T) {
	annsByFile := map[string]Annotations{
		"hello_world.txt": Annotations{
			{0, 5, []byte("<b>"), []byte("</b>"), 0},
			{7, 12, []byte("<i>"), []byte("</i>"), 0},
		},
		"adjacent.txt": Annotations{
			{0, 3, []byte("<b>"), []byte("</b>"), 0},
			{3, 6, []byte("<i>"), []byte("</i>"), 0},
		},
		"nested_0.txt": Annotations{
			{0, 4, []byte("<1>"), []byte("</1>"), 0},
			{1, 3, []byte("<2>"), []byte("</2>"), 0},
		},
		"nested_2.txt": Annotations{
			{0, 2, []byte("<1>"), []byte("</1>"), 0},
			{2, 4, []byte("<2>"), []byte("</2>"), 0},
			{4, 6, []byte("<3>"), []byte("</3>"), 0},
			{7, 8, []byte("<4>"), []byte("</4>"), 0},
		},
		"html.txt": Annotations{
			{193, 203, []byte("<1>"), []byte("</1>"), 0},
			{336, 339, []byte("<WOOF>"), []byte("</WOOF>"), 0},
		},
	}

	dir := "testdata"
	tests, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		name := test.Name()
		if !strings.Contains(name, *match) {
			continue
		}
		if strings.HasSuffix(name, ".html") {
			continue
		}
		path := filepath.Join(dir, name)
		input, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
			continue
		}

		anns := annsByFile[name]
		sort.Sort(anns)

		got, err := Annotate(input, anns, func(w io.Writer, b []byte) { template.HTMLEscape(w, b) })
		if err != nil {
			t.Errorf("%s: Annotate: %s", name, err)
			continue
		}

		expPath := path + ".html"
		if *saveExp {
			err = ioutil.WriteFile(expPath, got, 0700)
			if err != nil {
				t.Fatal(err)
			}
			continue
		}

		want, err := ioutil.ReadFile(expPath)
		if err != nil {
			t.Fatal(err)
		}

		want = bytes.TrimSpace(want)
		got = bytes.TrimSpace(got)

		if !bytes.Equal(want, got) {
			t.Errorf("%s: want %q, got %q", name, want, got)
			continue
		}
	}

	if *saveExp {
		t.Fatal("overwrote all expected output files with actual output (run tests again without -exp)")
	}
}

func BenchmarkAnnotate(b *testing.B) {
	input := []byte(strings.Repeat(strings.Repeat("a", 99)+"⌘", 20))
	inputLength := utf8.RuneCount(input)
	n := len(input)/2 - 50
	anns := make(Annotations, n)
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			anns[i] = &Annotation{Start: 2 * i, End: 2*i + 1}
		} else {
			anns[i] = &Annotation{Start: 2*i - 50, End: 2*i + 50}
			if anns[i].Start < 0 {
				anns[i].Start = 0
				anns[i].End = i
			}
			if anns[i].End >= inputLength {
				anns[i].End = inputLength
			}
		}
		anns[i].Left = []byte(strings.Repeat("L", i%20))
		anns[i].Right = []byte(strings.Repeat("R", i%20))
		anns[i].WantInner = i % 5
	}
	sort.Sort(anns)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Annotate(input, anns, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
