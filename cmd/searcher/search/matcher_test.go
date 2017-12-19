package search

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp/syntax"
	"strconv"
	"testing"
	"testing/iotest"
	"testing/quick"

	"github.com/sourcegraph/lazyzip"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searcher/protocol"
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
	ar, err := githubStore.openReader(ctx, p.Repo, p.Commit)
	if err != nil {
		b.Fatal(err)
	}
	defer ar.Close()

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, _, err := concurrentFind(ctx, rg, ar.Reader(), 0)
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
		"foo":           "foo",
		"FoO":           "foo",
		"(?m:^foo)":     "(?m:^)foo", // regex parse simplifies to this
		"(?m:^FoO)":     "(?m:^)foo",
		"[A-Z]":         "[a-z]",
		"[^A-Z]":        "[^A-Z]", // before we matched lowercase, and still after
		"[Z]":           "z",
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
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var buf bytes.Buffer
			for i := 0; i < test.size; i++ {
				if _, err := buf.WriteString("a"); err != nil {
					t.Fatal(err)
				}
			}
			matches, limitHit, err := rg.Find(&buf)
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
	zr, err := lazyzip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}

	rg, err := compile(&protocol.PatternInfo{Pattern: pattern})
	if err != nil {
		t.Fatal(err)
	}
	fileMatches, limitHit, err := concurrentFind(context.Background(), rg, zr, 0)
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

// githubStore fetches from github and caches across test runs.
var githubStore = &Store{
	FetchTar: fetchTarFromGithub,
	Path:     "/tmp/search_test/store",
}

func init() {
	// Clear out store so we pick up changes in our store writing code.
	os.RemoveAll(githubStore.Path)
}

func fetchTarFromGithub(ctx context.Context, repo, rev string) (io.ReadCloser, error) {
	// key is a sha256 hash since we want to use it for the disk name
	h := sha256.Sum256([]byte(repo + " " + rev))
	key := hex.EncodeToString(h[:])
	path := filepath.Join("/tmp/search_test/codeload/", key+".tar.gz")

	// Check codeload cache first
	r, err := openGzipReader(path)
	if err == nil {
		return r, nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}

	// Fetch archive to a temporary path
	tmpPath := path + ".part"
	url := fmt.Sprintf("https://codeload.%s/tar.gz/%s", repo, rev)
	fmt.Println("fetching", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github repo archive: URL %s returned HTTP %d", url, resp.StatusCode)
	}
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer func() { os.Remove(tmpPath) }()
	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		return nil, err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return nil, err
	}

	return openGzipReader(path)
}

func openGzipReader(name string) (io.ReadCloser, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	r, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	return &gzipReadCloser{f: f, r: r}, nil
}

type gzipReadCloser struct {
	f *os.File
	r *gzip.Reader
}

func (z *gzipReadCloser) Read(p []byte) (int, error) {
	return z.r.Read(p)
}
func (z *gzipReadCloser) Close() error {
	err := z.r.Close()
	if err1 := z.f.Close(); err == nil {
		err = err1
	}
	return err
}
