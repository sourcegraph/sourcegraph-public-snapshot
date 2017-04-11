package search

import (
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
	"testing"
)

// go test -bench=. -benchtime=10s -run=^$ ./cmd/searcher/search

func BenchmarkBytesToLowerASCII(b *testing.B) {
	src := []byte("\tThe Quick Brown Fox juMPs over the LAZY dog!?")
	dst := make([]byte, len(src))
	for n := 0; n < b.N; n++ {
		bytesToLowerASCII(dst, src)
	}
}

func BenchmarkConcurrentFind_large_fixed(b *testing.B) {
	benchConcurrentFind(b, &Params{
		Repo:    "github.com/golang/go",
		Commit:  "0ebaca6ba27534add5930a95acffa9acff182e2b",
		Pattern: "error handler",
	})
}

func BenchmarkConcurrentFind_large_fixed_casesensitive(b *testing.B) {
	benchConcurrentFind(b, &Params{
		Repo:            "github.com/golang/go",
		Commit:          "0ebaca6ba27534add5930a95acffa9acff182e2b",
		Pattern:         "error handler",
		IsCaseSensitive: true,
	})
}

func BenchmarkConcurrentFind_small_fixed(b *testing.B) {
	benchConcurrentFind(b, &Params{
		Repo:    "github.com/sourcegraph/go-langserver",
		Commit:  "4193810334683f87b8ed5d896aa4753f0dfcdf20",
		Pattern: "object not found",
	})
}

func BenchmarkConcurrentFind_small_fixed_casesensitive(b *testing.B) {
	benchConcurrentFind(b, &Params{
		Repo:            "github.com/sourcegraph/go-langserver",
		Commit:          "4193810334683f87b8ed5d896aa4753f0dfcdf20",
		Pattern:         "object not found",
		IsCaseSensitive: true,
	})
}

func benchConcurrentFind(b *testing.B, p *Params) {
	if testing.Short() {
		b.Skip("")
	}
	b.ReportAllocs()

	err := validateParams(p)
	if err != nil {
		b.Fatal(err)
	}

	rg, err := compile(p)
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
		_, err := concurrentFind(ctx, rg, ar.Reader())
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
		"[A-Z]":         "[a-z]",
		"[^A-Z]":        "[^A-Z]", // before we matched lowercase, and still after
		"[Z]":           "z",
		"[abB-Z]":       "[b-za-b]",
		"([abB-Z]|FoO)": "([b-za-b]|foo)",
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
