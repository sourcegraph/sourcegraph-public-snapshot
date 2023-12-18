package search_test

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/grafana/regexp"
	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/zoekt"

	"github.com/sourcegraph/sourcegraph/internal/gitserver"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/internal/search"
	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type fileType int

const (
	typeFile fileType = iota
	typeSymlink
)

func TestSearch(t *testing.T) {
	// Create byte buffer of binary file
	miltonPNG := bytes.Repeat([]byte{0x00}, 32*1024)

	files := map[string]struct {
		body string
		typ  fileType
	}{
		"README.md": {`# Hello World

Hello world example in go`, typeFile},
		"file++.plus": {`filename contains regex metachars`, typeFile},
		"nonutf8.txt": {"file contains invalid utf8 \xC0 characters", typeFile},
		"main.go": {`package main

import "fmt"

func main() {
	fmt.Println("Hello world")
}
`, typeFile},
		"abc.txt":    {"w", typeFile},
		"milton.png": {string(miltonPNG), typeFile},
		"ignore.me":  {`func hello() string {return "world"}`, typeFile},
		"symlink":    {"abc.txt", typeSymlink},
	}

	cases := []struct {
		arg          protocol.PatternInfo
		contextLines int32
		want         autogold.Value
	}{{
		arg:  protocol.PatternInfo{Pattern: "foo"},
		want: autogold.Expect(""),
	}, {
		arg:  protocol.PatternInfo{Pattern: "World", IsCaseSensitive: true},
		want: autogold.Expect("README.md:1:1:\n# Hello World\n"),
	}, {
		arg: protocol.PatternInfo{Pattern: "world", IsCaseSensitive: true},
		want: autogold.Expect(`README.md:3:3:
Hello world example in go
main.go:6:6:
fmt.Println("Hello world")
`),
	}, {
		arg:          protocol.PatternInfo{Pattern: "world", IsCaseSensitive: true},
		contextLines: 1,
		want: autogold.Expect(`README.md:2:3:

Hello world example in go
main.go:5:7:
func main() {
fmt.Println("Hello world")
}
`),
	}, {
		arg:          protocol.PatternInfo{Pattern: "world", IsCaseSensitive: true},
		contextLines: 2,
		want: autogold.Expect(`README.md:1:3:
# Hello World

Hello world example in go
main.go:4:7:

func main() {
fmt.Println("Hello world")
}
`),
	}, {
		arg:          protocol.PatternInfo{Pattern: "world", IsCaseSensitive: true},
		contextLines: 999,
		want: autogold.Expect(`README.md:1:3:
# Hello World

Hello world example in go
main.go:1:7:
package main

import "fmt"

func main() {
fmt.Println("Hello world")
}
`),
	}, {
		arg: protocol.PatternInfo{Pattern: "world"},
		want: autogold.Expect(`README.md:1:1:
# Hello World
README.md:3:3:
Hello world example in go
main.go:6:6:
fmt.Println("Hello world")
`),
	}, {
		arg:  protocol.PatternInfo{Pattern: "func.*main"},
		want: autogold.Expect(""),
	}, {
		arg:  protocol.PatternInfo{Pattern: "func.*main", IsRegExp: true},
		want: autogold.Expect("main.go:5:5:\nfunc main() {\n"),
	}, {
		// https://github.com/sourcegraph/sourcegraph/issues/8155
		arg:  protocol.PatternInfo{Pattern: "^func", IsRegExp: true},
		want: autogold.Expect("main.go:5:5:\nfunc main() {\n"),
	}, {
		arg:  protocol.PatternInfo{Pattern: "^FuNc", IsRegExp: true},
		want: autogold.Expect("main.go:5:5:\nfunc main() {\n"),
	}, {
		arg:  protocol.PatternInfo{Pattern: "mai", IsWordMatch: true},
		want: autogold.Expect(""),
	}, {
		arg: protocol.PatternInfo{Pattern: "main", IsWordMatch: true},
		want: autogold.Expect(`main.go:1:1:
package main
main.go:5:5:
func main() {
`),
	}, {
		// Ensure we handle CaseInsensitive regexp searches with
		// special uppercase chars in pattern.
		arg: protocol.PatternInfo{Pattern: `printL\B`, IsRegExp: true},
		want: autogold.Expect(`main.go:6:6:
fmt.Println("Hello world")
`),
	}, {
		arg: protocol.PatternInfo{Pattern: "world", ExcludePattern: "README.md"},
		want: autogold.Expect(`main.go:6:6:
fmt.Println("Hello world")
`),
	}, {
		arg: protocol.PatternInfo{Pattern: "world", IncludePatterns: []string{`\.md$`}},
		want: autogold.Expect(`README.md:1:1:
# Hello World
README.md:3:3:
Hello world example in go
`),
	}, {
		arg:  protocol.PatternInfo{Pattern: "w", IncludePatterns: []string{`\.(md|txt)$`, `\.txt$`}},
		want: autogold.Expect("abc.txt:1:1:\nw\n"),
	}, {
		arg: protocol.PatternInfo{Pattern: "world", ExcludePattern: "README\\.md"},
		want: autogold.Expect(`main.go:6:6:
fmt.Println("Hello world")
`),
	}, {
		arg: protocol.PatternInfo{Pattern: "world", IncludePatterns: []string{"\\.md"}},
		want: autogold.Expect(`README.md:1:1:
# Hello World
README.md:3:3:
Hello world example in go
`),
	}, {
		arg: protocol.PatternInfo{Pattern: "w", IncludePatterns: []string{"\\.(md|txt)", "README"}},
		want: autogold.Expect(`README.md:1:1:
# Hello World
README.md:3:3:
Hello world example in go
`),
	}, {
		arg: protocol.PatternInfo{Pattern: "world", IncludePatterns: []string{`\.(MD|go)$`}, PathPatternsAreCaseSensitive: true},
		want: autogold.Expect(`main.go:6:6:
fmt.Println("Hello world")
`),
	}, {
		arg:          protocol.PatternInfo{Pattern: "world", IncludePatterns: []string{`\.(MD|go)$`}, PathPatternsAreCaseSensitive: true},
		contextLines: 1,
		want: autogold.Expect(`main.go:5:7:
func main() {
fmt.Println("Hello world")
}
`),
	}, {
		arg:          protocol.PatternInfo{Pattern: "world", IncludePatterns: []string{`\.(MD|go)$`}, PathPatternsAreCaseSensitive: true},
		contextLines: 2,
		want: autogold.Expect(`main.go:4:7:

func main() {
fmt.Println("Hello world")
}
`),
	}, {
		arg: protocol.PatternInfo{Pattern: "world", IncludePatterns: []string{`\.(MD|go)`}, PathPatternsAreCaseSensitive: true},
		want: autogold.Expect(`main.go:6:6:
fmt.Println("Hello world")
`),
	}, {
		arg:  protocol.PatternInfo{Pattern: "doesnotmatch"},
		want: autogold.Expect(""),
	}, {
		arg:  protocol.PatternInfo{Pattern: "", IsRegExp: false, IncludePatterns: []string{"\\.png"}, PatternMatchesPath: true},
		want: autogold.Expect("milton.png\n"),
	}, {
		arg: protocol.PatternInfo{Pattern: "package main\n\nimport \"fmt\"", IsCaseSensitive: false, IsRegExp: true, PatternMatchesPath: true, PatternMatchesContent: true},
		want: autogold.Expect(`main.go:1:3:
package main

import "fmt"
`),
	}, {
		arg: protocol.PatternInfo{Pattern: "package main\n\\s*import \"fmt\"", IsCaseSensitive: false, IsRegExp: true, PatternMatchesPath: true, PatternMatchesContent: true},
		want: autogold.Expect(`main.go:1:3:
package main

import "fmt"
`),
	}, {
		arg:  protocol.PatternInfo{Pattern: "package main\n", IsCaseSensitive: false, IsRegExp: true, PatternMatchesPath: true, PatternMatchesContent: true},
		want: autogold.Expect("main.go:1:2:\npackage main\n\n"),
	}, {
		arg: protocol.PatternInfo{Pattern: "package main\n\\s*", IsCaseSensitive: false, IsRegExp: true, PatternMatchesPath: true, PatternMatchesContent: true},
		want: autogold.Expect(`main.go:1:3:
package main

import "fmt"
`),
	}, {
		arg:  protocol.PatternInfo{Pattern: "\nfunc", IsCaseSensitive: false, IsRegExp: true, PatternMatchesPath: true, PatternMatchesContent: true},
		want: autogold.Expect("main.go:4:5:\n\nfunc main() {\n"),
	}, {
		arg: protocol.PatternInfo{Pattern: "\n\\s*func", IsCaseSensitive: false, IsRegExp: true, PatternMatchesPath: true, PatternMatchesContent: true},
		want: autogold.Expect(`main.go:3:5:
import "fmt"

func main() {
`),
	}, {
		arg: protocol.PatternInfo{Pattern: "package main\n\nimport \"fmt\"\n\nfunc main\\(\\) {", IsCaseSensitive: false, IsRegExp: true, PatternMatchesPath: true, PatternMatchesContent: true},
		want: autogold.Expect(`main.go:1:5:
package main

import "fmt"

func main() {
`),
	}, {
		arg: protocol.PatternInfo{Pattern: "\n", IsCaseSensitive: false, IsRegExp: true, PatternMatchesPath: true, PatternMatchesContent: true},
		want: autogold.Expect(`README.md:1:3:
# Hello World

Hello world example in go
main.go:1:8:
package main

import "fmt"

func main() {
fmt.Println("Hello world")
}

`),
	}, {
		arg: protocol.PatternInfo{Pattern: "^$", IsRegExp: true},
		want: autogold.Expect(`README.md:2:2:

main.go:2:2:

main.go:4:4:

main.go:8:8:

milton.png:1:1:

`),
	}, {
		arg: protocol.PatternInfo{
			Pattern:         "filename contains regex metachars",
			IncludePatterns: []string{regexp.QuoteMeta("file++.plus")},
			IsStructuralPat: true,
			IsRegExp:        true, // To test for a regression, imply that IsStructuralPat takes precedence.
		},
		want: autogold.Expect(`file++.plus:1:1:
filename contains regex metachars
`),
	}, {
		arg: protocol.PatternInfo{Pattern: "World", IsNegated: true},
		want: autogold.Expect(`abc.txt
file++.plus
milton.png
nonutf8.txt
symlink
`),
	}, {
		arg: protocol.PatternInfo{Pattern: "World", IsCaseSensitive: true, IsNegated: true},
		want: autogold.Expect(`abc.txt
file++.plus
main.go
milton.png
nonutf8.txt
symlink
`),
	}, {
		arg: protocol.PatternInfo{Pattern: "fmt", IsNegated: true},
		want: autogold.Expect(`README.md
abc.txt
file++.plus
milton.png
nonutf8.txt
symlink
`),
	}, {
		arg:  protocol.PatternInfo{Pattern: "abc", PatternMatchesPath: true, PatternMatchesContent: true},
		want: autogold.Expect("abc.txt\nsymlink:1:1:\nabc.txt\n"),
	}, {
		arg:  protocol.PatternInfo{Pattern: "abc", PatternMatchesPath: false, PatternMatchesContent: true},
		want: autogold.Expect("symlink:1:1:\nabc.txt\n"),
	}, {
		arg:  protocol.PatternInfo{Pattern: "abc", PatternMatchesPath: true, PatternMatchesContent: false},
		want: autogold.Expect("abc.txt\n"),
	}, {
		arg: protocol.PatternInfo{Pattern: "utf8", PatternMatchesPath: false, PatternMatchesContent: true},
		want: autogold.Expect(`nonutf8.txt:1:1:
file contains invalid utf8 � characters
`),
	}}

	zoektURL := newZoekt(t, &zoekt.Repository{}, nil)
	s := newStore(t, files)
	s.FilterTar = func(_ context.Context, _ gitserver.Client, _ api.RepoName, _ api.CommitID) (search.FilterFunc, error) {
		return func(hdr *tar.Header) bool {
			return hdr.Name == "ignore.me"
		}, nil
	}
	ts := httptest.NewServer(&search.Service{
		Store:   s,
		Log:     s.Log,
		Indexed: backend.ZoektDial(zoektURL),
	})
	defer ts.Close()

	for i, test := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if test.arg.IsStructuralPat {
				maybeSkipComby(t)
			}

			req := protocol.Request{
				Repo:            "foo",
				URL:             "u",
				Commit:          "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
				PatternInfo:     test.arg,
				FetchTimeout:    fetchTimeoutForCI(t),
				NumContextLines: test.contextLines,
			}
			m, err := doSearch(ts.URL, &req)
			if err != nil {
				t.Fatalf("%s failed: %s", test.arg.String(), err)
			}
			sort.Sort(sortByPath(m))
			got := toString(m)
			err = sanityCheckSorted(m)
			if err != nil {
				t.Fatalf("%s malformed response: %s\n%s", test.arg.String(), err, got)
			}
			test.want.Equal(t, got)
		})
	}
}

func maybeSkipComby(t *testing.T) {
	t.Helper()
	if os.Getenv("CI") != "" {
		return
	}
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		t.Skip("Skipping due to limitations in comby and M1")
	}
	if _, err := exec.LookPath("comby"); err != nil {
		t.Skipf("skipping comby test when not on CI: %v", err)
	}
}

func TestSearch_badrequest(t *testing.T) {
	cases := []protocol.Request{
		// Bad regexp
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			PatternInfo: protocol.PatternInfo{
				Pattern:  `\F`,
				IsRegExp: true,
			},
		},

		// Unsupported regex
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			PatternInfo: protocol.PatternInfo{
				Pattern:  `(?!id)entity`,
				IsRegExp: true,
			},
		},

		// No repo
		{
			URL:    "u",
			Commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			PatternInfo: protocol.PatternInfo{
				Pattern: "test",
			},
		},

		// No commit
		{
			Repo: "foo",
			URL:  "u",
			PatternInfo: protocol.PatternInfo{
				Pattern: "test",
			},
		},

		// Non-absolute commit
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "HEAD",
			PatternInfo: protocol.PatternInfo{
				Pattern: "test",
			},
		},

		// Bad include glob
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			PatternInfo: protocol.PatternInfo{
				Pattern:         "test",
				IncludePatterns: []string{"[c-a]"},
			},
		},

		// Bad exclude glob
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			PatternInfo: protocol.PatternInfo{
				Pattern:        "test",
				ExcludePattern: "[c-a]",
			},
		},

		// Bad include regexp
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			PatternInfo: protocol.PatternInfo{
				Pattern:         "test",
				IncludePatterns: []string{"**"},
			},
		},

		// Bad exclude regexp
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			PatternInfo: protocol.PatternInfo{
				Pattern:        "test",
				ExcludePattern: "**",
			},
		},

		// structural search with negated pattern
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			PatternInfo: protocol.PatternInfo{
				Pattern:         "fmt.Println(:[_])",
				IsNegated:       true,
				ExcludePattern:  "",
				IsStructuralPat: true,
			},
		},
	}

	store := newStore(t, nil)
	ts := httptest.NewServer(&search.Service{
		Store: store,
		Log:   store.Log,
	})
	defer ts.Close()

	for _, p := range cases {
		p.PatternInfo.PatternMatchesContent = true
		_, err := doSearch(ts.URL, &p)
		if err == nil {
			t.Fatalf("%v expected to fail", p)
		}
	}
}

func doSearch(u string, p *protocol.Request) ([]protocol.FileMatch, error) {
	reqBody, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(u, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.Errorf("non-200 response: code=%d body=%s", resp.StatusCode, string(body))
	}

	var ed searcher.EventDone
	var matches []protocol.FileMatch
	dec := searcher.StreamDecoder{
		OnMatches: func(newMatches []*protocol.FileMatch) {
			for _, match := range newMatches {
				matches = append(matches, *match)
			}
		},
		OnDone: func(e searcher.EventDone) {
			ed = e
		},
		OnUnknown: func(event []byte, _ []byte) {
			panic("unknown event")
		},
	}
	if err := dec.ReadAll(resp.Body); err != nil {
		return nil, err
	}
	if ed.Error != "" {
		return nil, errors.New(ed.Error)
	}
	return matches, err
}

func newStore(t *testing.T, files map[string]struct {
	body string
	typ  fileType
},
) *search.Store {
	writeTar := func(w io.Writer, paths []string) error {
		if paths == nil {
			for name := range files {
				paths = append(paths, name)
			}
			sort.Strings(paths)
		}

		tarW := tar.NewWriter(w)
		for _, name := range paths {
			file := files[name]
			var hdr *tar.Header
			switch file.typ {
			case typeFile:
				hdr = &tar.Header{
					Name: name,
					Mode: 0o600,
					Size: int64(len(file.body)),
				}
				if err := tarW.WriteHeader(hdr); err != nil {
					return err
				}
				if _, err := tarW.Write([]byte(file.body)); err != nil {
					return err
				}
			case typeSymlink:
				hdr = &tar.Header{
					Typeflag: tar.TypeSymlink,
					Name:     name,
					Mode:     int64(os.ModePerm | os.ModeSymlink),
					Linkname: file.body,
				}
				if err := tarW.WriteHeader(hdr); err != nil {
					return err
				}
			}
		}
		// git-archive usually includes a pax header we should ignore.
		// use a body which matches a test case. Ensures we don't return this
		// false entry as a result.
		if err := addpaxheader(tarW, "Hello world\n"); err != nil {
			return err
		}

		return tarW.Close()
	}

	return &search.Store{
		GitserverClient: gitserver.NewTestClient(t),
		FetchTar: func(ctx context.Context, repo api.RepoName, commit api.CommitID) (io.ReadCloser, error) {
			r, w := io.Pipe()
			go func() {
				err := writeTar(w, nil)
				w.CloseWithError(err)
			}()
			return r, nil
		},
		FetchTarPaths: func(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) (io.ReadCloser, error) {
			r, w := io.Pipe()
			go func() {
				err := writeTar(w, paths)
				w.CloseWithError(err)
			}()
			return r, nil
		},
		Path: t.TempDir(),
		Log:  logtest.Scoped(t),

		ObservationCtx: observation.TestContextTB(t),
	}
}

// fetchTimeoutForCI gives a large timeout for CI. CI can be very busy, so we
// give a large timeout instead of giving bad signal on PRs.
func fetchTimeoutForCI(t *testing.T) time.Duration {
	if deadline, ok := t.Deadline(); ok {
		return time.Until(deadline) / 2
	}
	return 500 * time.Millisecond
}

func toString(m []protocol.FileMatch) string {
	buf := new(bytes.Buffer)
	for _, f := range m {
		if len(f.ChunkMatches) == 0 {
			buf.WriteString(f.Path)
			buf.WriteByte('\n')
		}
		for _, cm := range f.ChunkMatches {
			buf.WriteString(f.Path)
			buf.WriteByte(':')
			buf.WriteString(strconv.Itoa(int(cm.ContentStart.Line) + 1))
			buf.WriteByte(':')
			buf.WriteString(strconv.Itoa(int(cm.ContentStart.Line) + strings.Count(cm.Content, "\n") + 1))
			buf.WriteByte(':')
			buf.WriteByte('\n')
			buf.WriteString(cm.Content)
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

func sanityCheckSorted(m []protocol.FileMatch) error {
	if !sort.IsSorted(sortByPath(m)) {
		return errors.New("unsorted file matches, please sortByPath")
	}
	for i := range m {
		if i > 0 && m[i].Path == m[i-1].Path {
			return errors.Errorf("duplicate FileMatch on %s", m[i].Path)
		}
		cm := m[i].ChunkMatches
		if !sort.IsSorted(sortByLineNumber(cm)) {
			return errors.Errorf("unsorted LineMatches for %s", m[i].Path)
		}
		for j := range cm {
			if j > 0 && cm[j].ContentStart.Line == cm[j-1].ContentStart.Line {
				return errors.Errorf("duplicate LineNumber on %s:%d", m[i].Path, cm[j].ContentStart.Line)
			}
		}
	}
	return nil
}

type sortByPath []protocol.FileMatch

func (m sortByPath) Len() int           { return len(m) }
func (m sortByPath) Less(i, j int) bool { return m[i].Path < m[j].Path }
func (m sortByPath) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

type sortByLineNumber []protocol.ChunkMatch

func (m sortByLineNumber) Len() int           { return len(m) }
func (m sortByLineNumber) Less(i, j int) bool { return m[i].ContentStart.Line < m[j].ContentStart.Line }
func (m sortByLineNumber) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
