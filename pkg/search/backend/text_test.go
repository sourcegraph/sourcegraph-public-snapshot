package backend_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	zoektrpc "github.com/google/zoekt/rpc"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/endpoint"
	"github.com/sourcegraph/sourcegraph/pkg/search"
	"github.com/sourcegraph/sourcegraph/pkg/search/backend"
	"github.com/sourcegraph/sourcegraph/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/pkg/search/rpc"
)

func TestText(t *testing.T) {
	mz := &mockZoekt{SearchResult: &zoekt.SearchResult{}}
	addr1, close1 := openZoektServer(t, mz)
	defer close1()
	index := &backend.Zoekt{
		Client:       zoektrpc.Client(addr1),
		DisableCache: true,
	}
	defer index.Close()

	fallback := &mockCollectRepos{}
	addr2, close2 := openServer(t, fallback)
	defer close2()
	jit := &backend.TextJIT{
		Endpoints: endpoint.New(addr2),
		Resolve: func(ctx context.Context, name api.RepoName, spec string) (api.CommitID, error) {
			if spec == "" {
				spec = "HEAD"
			}
			return api.CommitID(strings.ToUpper(spec)), nil
		},
	}
	defer jit.Close()

	s := &backend.Text{
		Index:    index,
		Fallback: jit,
	}
	defer s.Close()

	// Test string codepath
	t.Log(s.String())

	cases := []struct {
		Name         string
		Query        string
		Repos        string
		Indexed      string
		WantIndex    string
		WantFallback string
		WantError    string
	}{{
		Name:      "all",
		Repos:     "a b c",
		Indexed:   "a b c",
		WantIndex: "a b c",
	}, {
		Name:         "none",
		Repos:        "a b c",
		Indexed:      "d e f",
		WantFallback: "a@HEAD b@HEAD c@HEAD",
	}, {
		Name:         "empty_index",
		Repos:        "a b c",
		WantFallback: "a@HEAD b@HEAD c@HEAD",
	}, {
		Name:      "subset_of_indexed",
		Repos:     "a b c",
		Indexed:   "a b c d",
		WantIndex: "a b c",
	}, {
		Name:         "query_has_extra_repos",
		Query:        "(r:a ref:x) or r:b",
		Repos:        "b",
		Indexed:      "a d e",
		WantFallback: "b@X",
	}, {
		// We send b to fallback since it matches b@x since searching b on the
		// x branch will return results.
		Name:         "query_has_extra_repos_indexed",
		Query:        "(r:a ref:x) or r:b",
		Repos:        "b",
		Indexed:      "b d e",
		WantFallback: "b@X",
	}, {
		Name:         "mix",
		Query:        "(r:a ref:x) or (r:b ref:y) or r:c or r:d",
		Repos:        "a b c d",
		Indexed:      "a d e",
		WantIndex:    "",
		WantFallback: "a@X b@Y c@X c@Y d@X d@Y",
	}, {
		Name:         "mix_on_same_repo",
		Query:        "(r:a ref:x) or (r:b ref:x)",
		Repos:        "a b c d",
		Indexed:      "a d e",
		WantFallback: "a@X b@X",
	}, {
		Name:         "no-head",
		Query:        "ref:x",
		Repos:        "a b c d",
		Indexed:      "a b c",
		WantFallback: "a@X b@X c@X d@X",
	}, {
		Name:      "empty",
		WantError: "repository list empty",
	}}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Reset mocks
			fallback.Repos = nil
			mz.LastSearchQ = nil
			mz.ListResult = zoektRepoList(c.Indexed)

			// Build a query which matches c.Repos
			var q query.Q
			if c.Query == "" {
				q = &query.Const{Value: true}
			} else {
				var err error
				q, err = query.Parse(c.Query)
				if err != nil {
					t.Fatal(q)
				}
				// Replace repo with reposet
				q, err = query.ExpandRepo(q, func(p, e []string) (map[string]struct{}, error) {
					if len(e) > 0 {
						return nil, errors.Errorf("excludes: %v", e)
					}
					m := map[string]struct{}{}
					for _, r := range p {
						m[r] = struct{}{}
					}
					return m, nil
				})
				if err != nil {
					t.Fatal(err)
				}
				t.Log(q)
			}
			_, err := s.Search(context.Background(), q, &search.Options{Repositories: repoList(c.Repos)})
			assertError(t, err, c.WantError)

			var gotIndexParts []string
			if mz.LastSearchQ != nil {
				zoektquery.VisitAtoms(mz.LastSearchQ, func(q zoektquery.Q) {
					if rs, ok := q.(*zoektquery.RepoSet); ok {
						for name := range rs.Set {
							seen := false
							for _, p := range gotIndexParts {
								if p == name {
									seen = true
								}
							}
							if !seen {
								gotIndexParts = append(gotIndexParts, name)
							}
						}
					}
				})
			}
			sort.Strings(gotIndexParts)
			gotIndex := strings.Join(gotIndexParts, " ")
			if gotIndex != c.WantIndex {
				t.Errorf("unexpected repos sent to index\ngot:  %s\nwant: %s", gotIndex, c.WantIndex)
			}

			var gotFallbackParts []string
			for _, r := range fallback.Repos {
				gotFallbackParts = append(gotFallbackParts, r.String())
			}
			sort.Strings(gotFallbackParts)
			gotFallback := strings.Join(gotFallbackParts, " ")
			if gotFallback != c.WantFallback {
				t.Errorf("unexpected repos sent to fallback\ngot:  %s\nwant: %s", gotFallback, c.WantFallback)
			}
		})
	}
}

func openServer(t *testing.T, s search.Searcher) (string, func()) {
	server, err := rpc.Server(s)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(server)
	return ts.URL, ts.Close
}

func openZoektServer(t *testing.T, s zoekt.Searcher) (string, func()) {
	server := zoektrpc.Server(s)
	ts := httptest.NewServer(server)
	return strings.TrimPrefix(ts.URL, "http://"), ts.Close
}

func TestText_error(t *testing.T) {
	mz := &mockZoekt{SearchResult: &zoekt.SearchResult{}}
	index := &backend.Zoekt{
		Client:       mz,
		DisableCache: true,
	}
	fallback := &backend.Mock{Result: &search.Result{}}
	s := &backend.Text{
		Index:    index,
		Fallback: fallback,
	}

	defer fallback.Close()
	defer index.Close()
	defer s.Close()

	expectError := func(e string) {
		_, err := s.Search(context.Background(), &query.Const{Value: true}, &search.Options{Repositories: repoList("a b")})
		assertError(t, err, e)
	}

	// If we fail to list, then we should just skip index
	mz.ListError = errors.New("foo1")
	expectError("")
	if mz.LastSearchQ != nil {
		t.Fatal("expected not to search zoekt")
	}
	if fallback.LastOpts == nil {
		t.Fatal("expected to search jit")
	}

	fallback.LastOpts = nil
	mz.LastSearchQ = nil
	mz.ListError = nil
	mz.ListResult = zoektRepoList("a")
	mz.SearchError = errors.New("foo2")
	expectError("foo2")

	mz.SearchError = nil
	fallback.Error = errors.New("foo3")
	expectError("foo3")
}

func TestZoekt(t *testing.T) {
	parse := func(s string) query.Q {
		q, err := query.Parse(s)
		if err != nil {
			t.Fatal(s, err)
		}
		return query.Simplify(q)
	}

	mock := &mockZoekt{
		ListResult: zoektRepoList("a b c"),
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
		WantError  string

		ZoektResult   *zoekt.SearchResult
		ZoektError    error
		WantZoektQ    string
		WantZoektOpts *zoekt.SearchOptions
	}{{
		Name: "simple",
		Q:    parse("foo"),
		Opts: &search.Options{Repositories: repoList("a")},

		ZoektResult: &zoekt.SearchResult{},
		WantZoektQ:  `(and (reposet a) substr:"foo")`,
	}, {
		Name: "ref",
		Q:    parse("(foo ref:x) or bar"),
		Opts: &search.Options{Repositories: repoList("a")},

		ZoektResult: &zoekt.SearchResult{},
		WantZoektQ:  `(and (reposet a) substr:"bar")`,
	}, {
		Name: "complex",
		Q:    query.NewAnd(parse("type:file -foo$"), query.NewRepoSet("a")),
		Opts: &search.Options{Repositories: repoList("a")},

		ZoektResult: &zoekt.SearchResult{},
		WantZoektQ:  `(and (reposet a) (type:filematch (not regex:"foo(?m:$)")) (reposet a))`,
	}, {
		Name: "error-repo",
		Q:    parse("foo r:a"),
		Opts: &search.Options{Repositories: repoList("a")},

		WantError: "zoekt does not allow repo atom",
	}, {
		Name: "error-empty",
		Q:    parse("foo"),
		Opts: &search.Options{},

		WantError: "repository list empty",
	}, {
		Name: "results",
		Q:    parse("foo"),
		Opts: &search.Options{Repositories: repoList("a")},
		WantResult: `
a:src/do_foo.go
a:src/do_test.go:6:func foo() {
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
			if assertError(t, err, c.WantError) {
				return
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

	if m.SearchResult == nil && m.SearchError == nil {
		return nil, errors.Errorf("mockZoekt.Search not set for %v", q)
	}

	return m.SearchResult, m.SearchError
}

func (m *mockZoekt) List(ctx context.Context, q zoektquery.Q) (*zoekt.RepoList, error) {
	m.LastListQ = q

	if m.ListResult == nil && m.ListError == nil {
		return nil, errors.Errorf("mockZoekt.List not set for %v", q)
	}

	return m.ListResult, m.ListError
}

func (m *mockZoekt) Close() {}
func (m *mockZoekt) String() string {
	return "mockZoekt"
}

func zoektRepoList(names string) *zoekt.RepoList {
	var repos []*zoekt.RepoListEntry
	for _, name := range strings.Fields(names) {
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

func repoList(specs string) []api.RepoName {
	var res []api.RepoName
	for _, name := range strings.Fields(specs) {
		res = append(res, api.RepoName(name))
	}
	return res
}

func assertError(t *testing.T, err error, contains string) bool {
	t.Helper()
	if contains == "" {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return err != nil
	}

	if err == nil {
		t.Fatalf("expected error containing %q, got nil", contains)
	}

	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("expected error containing %q, got %q", contains, err.Error())
	}
	return err != nil
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

type mockCollectRepos struct {
	mu    sync.Mutex
	Repos []search.Repository
}

func (m *mockCollectRepos) Search(ctx context.Context, q query.Q, opts *search.Options) (*search.Result, error) {
	var commits []api.CommitID
	var patterns []string
	query.VisitAtoms(q, func(q query.Q) {
		if s, ok := q.(*query.Ref); ok {
			if len(s.Pattern) == 40 {
				commits = append(commits, api.CommitID(s.Pattern))
			} else {
				patterns = append(patterns, s.Pattern)
			}
		}
	})
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, name := range opts.Repositories {
		if len(commits) == 0 && len(patterns) == 0 {
			m.Repos = append(m.Repos, search.Repository{
				Name: name,
			})
		}
		for _, c := range commits {
			m.Repos = append(m.Repos, search.Repository{
				Name:   name,
				Commit: c,
			})
		}
		for _, p := range patterns {
			m.Repos = append(m.Repos, search.Repository{
				Name:       name,
				RefPattern: p,
			})
		}
	}
	return &search.Result{}, nil
}

func (m *mockCollectRepos) Close() {}

func (m *mockCollectRepos) String() string { return "mockCollectRepos" }
