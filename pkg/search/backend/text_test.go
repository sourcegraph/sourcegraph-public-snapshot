package backend_test

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/search"
	"github.com/sourcegraph/sourcegraph/pkg/search/backend"
	"github.com/sourcegraph/sourcegraph/pkg/search/query"
)

func TestText(t *testing.T) {
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

	// Test string codepath
	t.Log(s.String())

	cases := []struct {
		Name         string
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
		WantFallback: "a b c",
	}, {
		Name:         "empty_index",
		Repos:        "a b c",
		WantFallback: "a b c",
	}, {
		Name:      "subset_of_indexed",
		Repos:     "a b c",
		Indexed:   "a b c d",
		WantIndex: "a b c",
	}, {
		Name:         "mix",
		Repos:        "a@x b@x c d",
		Indexed:      "a d e",
		WantIndex:    "d",
		WantFallback: "a@x b@x c",
	}, {
		Name:         "no-head",
		Repos:        "a@x b@x c@x d@x",
		Indexed:      "a b c",
		WantFallback: "a@x b@x c@x d@x",
	}, {
		Name:      "empty",
		WantError: "repository list empty",
	}}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Reset mocks
			fallback.LastOpts = nil
			mz.LastSearchQ = nil
			mz.ListResult = zoektRepoList(c.Indexed)

			_, err := s.Search(context.Background(), &query.Const{Value: true}, &search.Options{Repositories: repoList(c.Repos)})
			assertError(t, err, c.WantError)

			var gotIndexParts []string
			if mz.LastSearchQ != nil {
				zoektquery.VisitAtoms(mz.LastSearchQ, func(q zoektquery.Q) {
					if rs, ok := q.(*zoektquery.RepoSet); ok {
						for name := range rs.Set {
							gotIndexParts = append(gotIndexParts, name)
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
			if fallback.LastOpts != nil {
				for _, r := range fallback.LastOpts.Repositories {
					gotFallbackParts = append(gotFallbackParts, r.String())
				}
			}
			gotFallback := strings.Join(gotFallbackParts, " ")
			if gotFallback != c.WantFallback {
				t.Errorf("unexpected repos sent to index\ngot:  %s\nwant: %s", gotFallback, c.WantFallback)
			}
		})
	}
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

func repoList(specs string) []search.Repository {
	var res []search.Repository
	for _, spec := range strings.Fields(specs) {
		p := strings.Split(spec, "@")
		if len(p) == 1 {
			res = append(res, search.Repository{
				Name: api.RepoName(p[0]),
			})
		} else {
			res = append(res, search.Repository{
				Name:   api.RepoName(p[0]),
				Commit: api.CommitID(p[1]),
			})
		}
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
