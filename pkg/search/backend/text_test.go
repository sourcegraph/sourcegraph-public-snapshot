package backend_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/sourcegraph/sourcegraph/pkg/search"
	"github.com/sourcegraph/sourcegraph/pkg/search/backend"
	"github.com/sourcegraph/sourcegraph/pkg/search/query"
)

func TestZoekt(t *testing.T) {
	parse := func(s string) query.Q {
		q, err := query.Parse(s)
		if err != nil {
			t.Fatal(s, err)
		}
		return query.Simplify(q)
	}

	mock := &mockZoekt{
		ListResult: zoektRepoList("a", "b", "c"),
	}
	z := &backend.Zoekt{
		Client:       mock,
		DisableCache: true,
	}
	defer z.Close()

	cases := []struct {
		Name string

		Q          query.Q
		Opts       *search.Options
		WantResult string
		WantError  error

		ZoektResult   *zoekt.SearchResult
		ZoektError    error
		WantZoektQ    string
		WantZoektOpts *zoekt.SearchOptions
	}{{
		Name: "simple",
		Q:    parse("foo"),
		Opts: &search.Options{
			Repositories: []search.Repository{{Name: "a"}},
		},

		ZoektResult: &zoekt.SearchResult{},
		WantZoektQ:  `(and (reposet a) substr:"foo")`,
	}, {
		Name: "ref",
		Q:    parse("(foo ref:x) or bar"),
		Opts: &search.Options{
			Repositories: []search.Repository{{Name: "a"}},
		},

		ZoektResult: &zoekt.SearchResult{},
		WantZoektQ:  `(and (reposet a) substr:"bar")`,
	}, {
		Name: "results",
		Q:    parse("foo"),
		Opts: &search.Options{
			Repositories: []search.Repository{{Name: "a"}},
		},
		WantResult: `
a@:src/do_foo.go
a@:src/do_test.go:6:func foo() {
`,

		ZoektResult: &zoekt.SearchResult{
			Files: []zoekt.FileMatch{{
				FileName:   "src/do_foo.go",
				Repository: "a",
				LineMatches: []zoekt.LineMatch{{
					FileName: true,
				}},
			}, {
				FileName:   "src/do_test.go",
				Repository: "a",
				LineMatches: []zoekt.LineMatch{{
					Line:       []byte("func foo() {"),
					LineStart:  33,
					LineEnd:    45,
					LineNumber: 5,
					LineFragments: []zoekt.LineFragmentMatch{{
						LineOffset:  6,
						MatchLength: 3,
					}},
				}},
			}},
		},
		WantZoektQ: `(and (reposet a) substr:"foo")`,
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			mock.SearchResult = c.ZoektResult
			mock.SearchError = c.ZoektError

			results, err := z.Search(context.Background(), c.Q, c.Opts)
			if err != c.WantError {
				t.Fatalf("got error %v, want %v", err, c.WantError)
			}

			if c.WantZoektQ != "" && (mock.LastSearchQ.String() != c.WantZoektQ) {
				t.Errorf("unexpected query passed to zoekt:\ngot:  %v\nwant: %s", mock.LastSearchQ, c.WantZoektQ)
			}
			if c.WantZoektOpts != nil && !reflect.DeepEqual(mock.LastSearchOpts, c.WantZoektOpts) {
				t.Errorf("unexpected opts passed to zoekt:\ngot:  %#+v\nwant: %#+v", mock.LastSearchOpts, c.WantZoektOpts)
			}

			assertResults(t, results, strings.TrimLeft(c.WantResult, " \n"))
		})
	}
}

type mockZoekt struct {
	SearchResult   *zoekt.SearchResult
	SearchError    error
	LastSearchQ    zoektquery.Q
	LastSearchOpts *zoekt.SearchOptions

	ListResult *zoekt.RepoList
	ListError  error
	LastListQ  zoektquery.Q
}

func (m *mockZoekt) Search(ctx context.Context, q zoektquery.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	m.LastSearchQ = q
	m.LastSearchOpts = opts

	return m.SearchResult, m.SearchError
}

func (m *mockZoekt) List(ctx context.Context, q zoektquery.Q) (*zoekt.RepoList, error) {
	m.LastListQ = q

	return m.ListResult, m.ListError
}

func (m *mockZoekt) Close() {}
func (m *mockZoekt) String() string {
	return "mockZoekt"
}

func zoektRepoList(names ...string) *zoekt.RepoList {
	repos := make([]*zoekt.RepoListEntry, 0, len(names))
	for _, name := range names {
		repos = append(repos, &zoekt.RepoListEntry{
			Repository: zoekt.Repository{
				Name: name,
			},
		})
	}
	return &zoekt.RepoList{
		Repos: repos,
	}
}

func assertResults(t *testing.T, r *search.Result, want string) {
	t.Helper()
	var got string
	if r == nil {
		got = "nil"
	} else {
		var buf bytes.Buffer
		for _, f := range r.Files {
			buf.WriteString(f.Repository.String())
			buf.WriteByte(':')
			if len(f.LineMatches) == 0 {
				buf.WriteString(f.Path)
				buf.WriteByte('\n')
			}
			for _, l := range f.LineMatches {
				buf.WriteString(f.Path)
				buf.WriteByte(':')
				buf.WriteString(strconv.Itoa(l.LineNumber + 1))
				buf.WriteByte(':')
				buf.Write(l.Line)
				buf.WriteByte('\n')
			}
		}
		got = buf.String()
	}
	if got == want {
		return
	}
	d, err := diff(want, got)
	if err != nil {
		t.Fatal(err)
	}
	t.Fatalf("unexpected search result:\n%s", d)
}

func diff(b1, b2 string) (string, error) {
	f1, err := ioutil.TempFile("", "search_test")
	if err != nil {
		return "", err
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("", "search_test")
	if err != nil {
		return "", err
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	f1.WriteString(b1)
	f2.WriteString(b2)

	data, err := exec.Command("diff", "-u", "--label=want", f1.Name(), "--label=got", f2.Name()).CombinedOutput()
	if len(data) > 0 {
		err = nil
	}
	return string(data), err
}
