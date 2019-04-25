package search

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"reflect"
	"regexp"
	"regexp/syntax"
	"sort"
	"strconv"
	"testing"
	"testing/iotest"
	"testing/quick"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/store"
	"github.com/sourcegraph/sourcegraph/pkg/testutil"
)

func benchBytesToLower(b *testing.B, src []byte) {
	dst := make([]byte, len(src))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytesToLowerASCII(dst, src)
	}
}

func BenchmarkBytesToLowerASCII(b *testing.B) {
	b.Run("short", func(b *testing.B) { benchBytesToLower(b, []byte("a-z@[A-Z")) })
	b.Run("pangram", func(b *testing.B) { benchBytesToLower(b, []byte("\tThe Quick Brown Fox juMPs over the LAZY dog!?")) })
	long := bytes.Repeat([]byte{'A'}, 8*1024)
	b.Run("8k", func(b *testing.B) { benchBytesToLower(b, long) })
	b.Run("8k-misaligned", func(b *testing.B) { benchBytesToLower(b, long[1:]) })
}

func checkBytesToLower(t *testing.T, b []byte) {
	t.Helper()
	want := make([]byte, len(b))
	bytesToLowerASCIIgeneric(want, b)
	got := make([]byte, len(b))
	bytesToLowerASCII(got, b)
	if !bytes.Equal(want, got) {
		t.Errorf("bytesToLowerASCII(%q)=%q want %q", b, got, want)
	}
}

func TestBytesToLowerASCII(t *testing.T) {
	// @ and [ are special: '@'+1=='A' and 'Z'+1=='['
	t.Run("pangram", func(t *testing.T) {
		checkBytesToLower(t, []byte("\t[The Quick Brown Fox juMPs over the LAZY dog!?@"))
	})
	t.Run("short", func(t *testing.T) {
		checkBytesToLower(t, []byte("a-z@[A-Z"))
	})
	t.Run("quick", func(t *testing.T) {
		f := func(b []byte) bool {
			x := make([]byte, len(b))
			bytesToLowerASCIIgeneric(x, b)
			y := make([]byte, len(b))
			bytesToLowerASCII(y, b)
			return bytes.Equal(x, y)
		}
		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})
	t.Run("alignment", func(t *testing.T) {
		// The goal of this test is to make sure we don't write to any bytes
		// that don't belong to us.
		b := make([]byte, 96)
		c := make([]byte, 96)
		for i := 0; i < len(b); i++ {
			for j := i; j < len(b); j++ {
				// fill b with Ms and c with xs
				for k := range b {
					b[k] = 'M'
					c[k] = 'x'
				}
				// process a subslice of b
				bytesToLowerASCII(c[i:j], b[i:j])
				for k := range b {
					want := byte('m')
					if k < i || k >= j {
						want = 'x'
					}
					if want != c[k] {
						t.Errorf("bytesToLowerASCII bad byte using bounds [%d:%d] (len %d) at index %d, have %c want %c", i, j, len(c[i:j]), k, c[k], want)
					}
				}
			}

		}
	})
}

func BenchmarkConcurrentFind_large_fixed(b *testing.B) {
	benchConcurrentFind(b, &protocol.Request{
		Repo:   "github.com/golang/go",
		Commit: "0ebaca6ba27534add5930a95acffa9acff182e2b",
		PatternInfo: protocol.PatternInfo{
			Pattern: "error handler",
		},
	})
}

func BenchmarkConcurrentFind_large_fixed_casesensitive(b *testing.B) {
	benchConcurrentFind(b, &protocol.Request{
		Repo:   "github.com/golang/go",
		Commit: "0ebaca6ba27534add5930a95acffa9acff182e2b",
		PatternInfo: protocol.PatternInfo{
			Pattern:         "error handler",
			IsCaseSensitive: true,
		},
	})
}

func BenchmarkConcurrentFind_large_re_dotstar(b *testing.B) {
	benchConcurrentFind(b, &protocol.Request{
		Repo:   "github.com/golang/go",
		Commit: "0ebaca6ba27534add5930a95acffa9acff182e2b",
		PatternInfo: protocol.PatternInfo{
			Pattern:  ".*",
			IsRegExp: true,
		},
	})
}

func BenchmarkConcurrentFind_large_re_common(b *testing.B) {
	benchConcurrentFind(b, &protocol.Request{
		Repo:   "github.com/golang/go",
		Commit: "0ebaca6ba27534add5930a95acffa9acff182e2b",
		PatternInfo: protocol.PatternInfo{
			Pattern:         "func +[A-Z]",
			IsRegExp:        true,
			IsCaseSensitive: true,
		},
	})
}

func BenchmarkConcurrentFind_large_re_anchor(b *testing.B) {
	// TODO(keegan) PERF regex engine performs poorly since LiteralPrefix
	// is empty when ^. We can improve this by:
	// * Transforming the regex we use to prune a file to be more
	// performant/permissive.
	// * Searching for any literal (Rabin-Karp aka bytes.Index) or group
	// of literals (Aho-Corasick).
	benchConcurrentFind(b, &protocol.Request{
		Repo:   "github.com/golang/go",
		Commit: "0ebaca6ba27534add5930a95acffa9acff182e2b",
		PatternInfo: protocol.PatternInfo{
			Pattern:         "^func +[A-Z]",
			IsRegExp:        true,
			IsCaseSensitive: true,
		},
	})
}

func BenchmarkConcurrentFind_large_path(b *testing.B) {
	do := func(b *testing.B, content, path bool) {
		benchConcurrentFind(b, &protocol.Request{
			Repo:   "github.com/golang/go",
			Commit: "0ebaca6ba27534add5930a95acffa9acff182e2b",
			PatternInfo: protocol.PatternInfo{
				Pattern:               "http.*client",
				IsRegExp:              true,
				IsCaseSensitive:       true,
				PatternMatchesContent: content,
				PatternMatchesPath:    path,
			},
		})
	}
	b.Run("path only", func(b *testing.B) { do(b, false, true) })
	b.Run("content only", func(b *testing.B) { do(b, true, false) })
	b.Run("both path and content", func(b *testing.B) { do(b, true, true) })
}

func BenchmarkConcurrentFind_small_fixed(b *testing.B) {
	benchConcurrentFind(b, &protocol.Request{
		Repo:   "github.com/sourcegraph/go-langserver",
		Commit: "4193810334683f87b8ed5d896aa4753f0dfcdf20",
		PatternInfo: protocol.PatternInfo{
			Pattern: "object not found",
		},
	})
}

func BenchmarkConcurrentFind_small_fixed_casesensitive(b *testing.B) {
	benchConcurrentFind(b, &protocol.Request{
		Repo:   "github.com/sourcegraph/go-langserver",
		Commit: "4193810334683f87b8ed5d896aa4753f0dfcdf20",
		PatternInfo: protocol.PatternInfo{
			Pattern:         "object not found",
			IsCaseSensitive: true,
		},
	})
}

func BenchmarkConcurrentFind_small_re_dotstar(b *testing.B) {
	benchConcurrentFind(b, &protocol.Request{
		Repo:   "github.com/sourcegraph/go-langserver",
		Commit: "4193810334683f87b8ed5d896aa4753f0dfcdf20",
		PatternInfo: protocol.PatternInfo{
			Pattern:  ".*",
			IsRegExp: true,
		},
	})
}

func BenchmarkConcurrentFind_small_re_common(b *testing.B) {
	benchConcurrentFind(b, &protocol.Request{
		Repo:   "github.com/sourcegraph/go-langserver",
		Commit: "4193810334683f87b8ed5d896aa4753f0dfcdf20",
		PatternInfo: protocol.PatternInfo{
			Pattern:         "func +[A-Z]",
			IsRegExp:        true,
			IsCaseSensitive: true,
		},
	})
}

func BenchmarkConcurrentFind_small_re_anchor(b *testing.B) {
	benchConcurrentFind(b, &protocol.Request{
		Repo:   "github.com/sourcegraph/go-langserver",
		Commit: "4193810334683f87b8ed5d896aa4753f0dfcdf20",
		PatternInfo: protocol.PatternInfo{
			Pattern:         "^func +[A-Z]",
			IsRegExp:        true,
			IsCaseSensitive: true,
		},
	})
}

func benchConcurrentFind(b *testing.B, p *protocol.Request) {
	if testing.Short() {
		b.Skip("")
	}
	b.ReportAllocs()

	err := validateParams(p)
	if err != nil {
		b.Fatal(err)
	}

	rg, err := compile(&p.PatternInfo)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	path, err := githubStore.PrepareZip(ctx, p.GitserverRepo(), p.Commit)
	if err != nil {
		b.Fatal(err)
	}

	var zc store.ZipCache
	zf, err := zc.Get(path)
	if err != nil {
		b.Fatal(err)
	}
	defer zf.Close()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, _, err := concurrentFind(ctx, rg, zf, 0, p.PatternMatchesContent, p.PatternMatchesPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestLowerRegexp(t *testing.T) {
	// The expected values are a bit volatile, since they come from
	// syntex.Regexp.String. So they may change between go versions. Just
	// ensure they make sense.
	cases := map[string]string{
		"foo":       "foo",
		"FoO":       "foo",
		"(?m:^foo)": "(?m:^)foo", // regex parse simplifies to this
		"(?m:^FoO)": "(?m:^)foo",

		// Ranges for the characters can be tricky. So we include many
		// cases. Importantly user intention when they write [^A-Z] is would
		// expect [^a-z] to apply when ignoring case.
		"[A-Z]":  "[a-z]",
		"[^A-Z]": "[^A-Za-z]",
		"[A-M]":  "[a-m]",
		"[^A-M]": "[^A-Ma-m]",
		"[A]":    "a",
		"[^A]":   "[^Aa]",
		"[M]":    "m",
		"[^M]":   "[^Mm]",
		"[Z]":    "z",
		"[^Z]":   "[^Zz]",
		"[a-z]":  "[a-z]",
		"[^a-z]": "[^a-z]",
		"[a-m]":  "[a-m]",
		"[^a-m]": "[^a-m]",
		"[a]":    "a",
		"[^a]":   "[^a]",
		"[m]":    "m",
		"[^m]":   "[^m]",
		"[z]":    "z",
		"[^z]":   "[^z]",

		// @ is tricky since it is 1 value less than A
		"[^A-Z@]": "[^@-Za-z]",

		// full unicode range should just be a .
		"[\\x00-\\x{10ffff}]": "(?s:.)",

		"[abB-Z]":       "[b-za-b]",
		"([abB-Z]|FoO)": "([b-za-b]|foo)",
		`[@-\[]`:        `[@-\[a-z]`,      // original range includes A-Z but excludes a-z
		`\S`:            `[^\t-\n\f-\r ]`, // \S is shorthand for the expected
	}

	for expr, want := range cases {
		re, err := syntax.Parse(expr, syntax.Perl)
		if err != nil {
			t.Fatal(expr, err)
		}
		lowerRegexpASCII(re)
		got := re.String()
		if want != got {
			t.Errorf("lowerRegexp(%q) == %q != %q", expr, got, want)
		}
	}
}

func TestLongestLiteral(t *testing.T) {
	cases := map[string]string{
		"foo":       "foo",
		"FoO":       "FoO",
		"(?m:^foo)": "foo",
		"(?m:^FoO)": "FoO",
		"[Z]":       "Z",

		`\wddSuballocation\(dump`:    "ddSuballocation(dump",
		`\wfoo(\dlongest\wbam)\dbar`: "longest",

		`(foo\dlongest\dbar)`:  "longest",
		`(foo\dlongest\dbar)+`: "longest",
		`(foo\dlongest\dbar)*`: "",

		"(foo|bar)":     "",
		"[A-Z]":         "",
		"[^A-Z]":        "",
		"[abB-Z]":       "",
		"([abB-Z]|FoO)": "",
		`[@-\[]`:        "",
		`\S`:            "",
	}

	metaLiteral := "AddSuballocation(dump->guid(), system_allocator_name)"
	cases[regexp.QuoteMeta(metaLiteral)] = metaLiteral

	for expr, want := range cases {
		re, err := syntax.Parse(expr, syntax.Perl)
		if err != nil {
			t.Fatal(expr, err)
		}
		re = re.Simplify()
		got := longestLiteral(re)
		if want != got {
			t.Errorf("longestLiteral(%q) == %q != %q", expr, got, want)
		}
	}
}

func TestReadAll(t *testing.T) {
	input := []byte("Hello World")

	// If we are the same size as input, it should work
	b := make([]byte, len(input))
	n, err := readAll(bytes.NewReader(input), b)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(input) {
		t.Fatalf("want to read in %d bytes, read %d", len(input), n)
	}
	if string(b[:n]) != string(input) {
		t.Fatalf("got %s, want %s", string(b[:n]), string(input))
	}

	// If we are larger then it should work
	b = make([]byte, len(input)*2)
	n, err = readAll(bytes.NewReader(input), b)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(input) {
		t.Fatalf("want to read in %d bytes, read %d", len(input), n)
	}
	if string(b[:n]) != string(input) {
		t.Fatalf("got %s, want %s", string(b[:n]), string(input))
	}

	// Same size, but modify reader to return 1 byte per call to ensure
	// our loop works.
	b = make([]byte, len(input))
	n, err = readAll(iotest.OneByteReader(bytes.NewReader(input)), b)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(input) {
		t.Fatalf("want to read in %d bytes, read %d", len(input), n)
	}
	if string(b[:n]) != string(input) {
		t.Fatalf("got %s, want %s", string(b[:n]), string(input))
	}

	// If we are too small it should fail
	b = make([]byte, 1)
	_, err = readAll(bytes.NewReader(input), b)
	if err == nil {
		t.Fatal("expected to fail on small buffer")
	}
}

func TestLineLimit(t *testing.T) {
	rg, err := compile(&protocol.PatternInfo{Pattern: "a"})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		size    int
		matches bool
	}{
		{size: maxLineSize, matches: true},
		{size: maxLineSize + 1, matches: false},
	}

	// calculate maximum size in tests,
	// needed because readerGreps re-use their buffers.
	maxBuf := 0
	for _, test := range tests {
		if test.size > maxBuf {
			maxBuf = test.size
		}
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			fakeZipFile := store.ZipFile{
				MaxLen: maxBuf,
				Data:   bytes.Repeat([]byte("A"), test.size),
			}
			fakeSrcFile := store.SrcFile{Len: int32(test.size)}
			matches, limitHit, err := rg.Find(&fakeZipFile, &fakeSrcFile)
			if err != nil {
				t.Fatal(err)
			}
			if limitHit {
				t.Fatalf("expected limit to not hit")
			}
			hasMatches := len(matches) != 0
			if hasMatches != test.matches {
				t.Fatalf("hasMatches=%t test.matches=%t", hasMatches, test.matches)
			}
		})
	}
}

func TestMaxMatches(t *testing.T) {
	pattern := "foo"

	// Create a zip archive which contains our limits + 1
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for i := 0; i < maxFileMatches+1; i++ {
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:   strconv.Itoa(i),
			Method: zip.Store,
		})
		if err != nil {
			t.Fatal(err)
		}
		for j := 0; j < maxLineMatches+1; j++ {
			for k := 0; k < maxOffsets+1; k++ {
				w.Write([]byte(pattern))
				w.Write([]byte{' '})
			}
			w.Write([]byte{'\n'})
		}
	}
	err := zw.Close()
	if err != nil {
		t.Fatal(err)
	}
	zf, err := store.MockZipFile(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	rg, err := compile(&protocol.PatternInfo{Pattern: pattern})
	if err != nil {
		t.Fatal(err)
	}
	fileMatches, limitHit, err := concurrentFind(context.Background(), rg, zf, 0, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if !limitHit {
		t.Fatalf("expected limitHit on concurrentFind")
	}

	if len(fileMatches) != maxFileMatches {
		t.Fatalf("expected %d file matches, got %d", maxFileMatches, len(fileMatches))
	}
	for _, fm := range fileMatches {
		if !fm.LimitHit {
			t.Fatalf("expected limitHit on file match")
		}
		if len(fm.LineMatches) != maxLineMatches {
			t.Fatalf("expected %d line matches, got %d", maxLineMatches, len(fm.LineMatches))
		}
		for _, lm := range fm.LineMatches {
			if !lm.LimitHit {
				t.Fatalf("expected limitHit on line match")
			}
			if len(lm.OffsetAndLengths) != maxOffsets {
				t.Fatalf("expected %d offsets, got %d", maxOffsets, len(lm.OffsetAndLengths))
			}
		}
	}
}

// Tests that:
//
// - IncludePatterns can match the path in any order
// - A path must match all (not any) of the IncludePatterns
// - An empty pattern is allowed
func TestPathMatches(t *testing.T) {
	zipData, err := createZip(map[string]string{
		"a":   "",
		"a/b": "",
		"a/c": "",
		"ab":  "",
		"b/a": "",
		"ba":  "",
		"c/d": "",
	})
	if err != nil {
		t.Fatal(err)
	}
	zf, err := store.MockZipFile(zipData)
	if err != nil {
		t.Fatal(err)
	}

	rg, err := compile(&protocol.PatternInfo{
		Pattern:                "",
		IncludePatterns:        []string{"a", "b"},
		PathPatternsAreRegExps: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	fileMatches, _, err := concurrentFind(context.Background(), rg, zf, 10, true, true)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"a/b", "ab", "b/a", "ba"}
	got := make([]string, len(fileMatches))
	for i, fm := range fileMatches {
		got[i] = fm.Path
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got file matches %v, want %v", got, want)
	}
}

func createZip(files map[string]string) ([]byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for name, body := range files {
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:   name,
			Method: zip.Store,
		})
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(body)); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil

}

// githubStore fetches from github and caches across test runs.
var githubStore = &store.Store{
	FetchTar: testutil.FetchTarFromGithub,
	Path:     "/tmp/search_test/store",
}

func init() {
	// Clear out store so we pick up changes in our store writing code.
	os.RemoveAll(githubStore.Path)
}
