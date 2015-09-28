package search

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph/mock"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/svc"
)

func TestNewPlan(t *testing.T) {
	tests := []struct {
		tokens  []sourcegraph.Token
		want    *sourcegraph.Plan
		wantErr error

		mockReposList func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error)
	}{
		{
			tokens: []sourcegraph.Token{sourcegraph.Term("a")},
			want: &sourcegraph.Plan{
				Repos: &sourcegraph.RepoListOptions{Query: "a", NoFork: true, Sort: "updated", Direction: "desc"},
				Users: &sourcegraph.UsersListOptions{Query: "a"},
				Defs:  &sourcegraph.DefListOptions{Query: "a", Nonlocal: true},
			},
		},
		{
			tokens: []sourcegraph.Token{sourcegraph.Term("a"), sourcegraph.Term("b")},
			want: &sourcegraph.Plan{
				Repos: &sourcegraph.RepoListOptions{Query: "a b", NoFork: true, Sort: "updated", Direction: "desc"},
				Users: &sourcegraph.UsersListOptions{Query: "a b"},
				Defs:  &sourcegraph.DefListOptions{Query: "a b", Nonlocal: true},
			},
		},
		{
			tokens: []sourcegraph.Token{sourcegraph.Term("a b")},
			want: &sourcegraph.Plan{
				Repos: &sourcegraph.RepoListOptions{Query: "a b", NoFork: true, Sort: "updated", Direction: "desc"},
				Users: &sourcegraph.UsersListOptions{Query: "a b"},
				Defs:  &sourcegraph.DefListOptions{Query: "a b", Nonlocal: true},
			},
		},
		{
			tokens: []sourcegraph.Token{sourcegraph.Term("x"), sourcegraph.Term("a b"), sourcegraph.Term("z")},
			want: &sourcegraph.Plan{
				Repos: &sourcegraph.RepoListOptions{Query: "x a b z", NoFork: true, Sort: "updated", Direction: "desc"},
				Users: &sourcegraph.UsersListOptions{Query: "x a b z"},
				Defs:  &sourcegraph.DefListOptions{Query: "x a b z", Nonlocal: true},
			},
		},
		{
			tokens: []sourcegraph.Token{sourcegraph.RepoToken{URI: "r"}},
			want: &sourcegraph.Plan{
				Repos: &sourcegraph.RepoListOptions{URIs: []string{"r"}, NoFork: true, Sort: "updated", Direction: "desc"},
				Defs:  &sourcegraph.DefListOptions{RepoRevs: []string{"r"}, Nonlocal: true},
			},
		},
		{
			tokens: []sourcegraph.Token{sourcegraph.RepoToken{URI: "r"}, sourcegraph.RevToken{Rev: "v"}},
			want: &sourcegraph.Plan{
				Repos: &sourcegraph.RepoListOptions{URIs: []string{"r"}, NoFork: true, Sort: "updated", Direction: "desc"},
				Defs:  &sourcegraph.DefListOptions{RepoRevs: []string{"r@v"}, Nonlocal: true},
			},
		},
		{
			tokens: []sourcegraph.Token{sourcegraph.RepoToken{URI: "r"}, sourcegraph.RevToken{Rev: ""}},
			want: &sourcegraph.Plan{
				Repos: &sourcegraph.RepoListOptions{URIs: []string{"r"}, NoFork: true, Sort: "updated", Direction: "desc"},
				Defs:  &sourcegraph.DefListOptions{RepoRevs: []string{"r@"}, Nonlocal: true},
			},
		},
		{
			tokens: []sourcegraph.Token{sourcegraph.RepoToken{URI: "r"}, sourcegraph.Term("t")},
			want: &sourcegraph.Plan{
				Repos:        &sourcegraph.RepoListOptions{Query: "t", URIs: []string{"r"}, NoFork: true, Sort: "updated", Direction: "desc"},
				Defs:         &sourcegraph.DefListOptions{Query: "t", RepoRevs: []string{"r"}, Nonlocal: true},
				TreeRepoRevs: []string{"r"},
				Tree: &sourcegraph.RepoTreeSearchOptions{
					SearchOptions: vcs.SearchOptions{
						Query:     "t",
						QueryType: vcs.FixedQuery,
						N:         10,
					},
				},
			},
		},
		{
			// TODO(sqs): implement this. somehow we need to expand
			// the RepoRevs list of u's repositories for a defs query.
			tokens: []sourcegraph.Token{sourcegraph.UserToken{Login: "u"}},
			want: &sourcegraph.Plan{
				Repos: &sourcegraph.RepoListOptions{Owner: "u", NoFork: true, Sort: "updated", Direction: "desc"},
				Users: &sourcegraph.UsersListOptions{Query: "u"},
				Defs: &sourcegraph.DefListOptions{
					RepoRevs: []string{"rr@a"},
					Nonlocal: true,
				},
			},
			mockReposList: func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
				return &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "rr", DefaultBranch: "a"}}}, nil
			},
		},

		{
			tokens: []sourcegraph.Token{sourcegraph.UserToken{Login: ""}},
			want:   &sourcegraph.Plan{},
		},

		// repo:@u expands to all repos owned by u.
		{
			tokens: []sourcegraph.Token{sourcegraph.UserToken{Login: "u"}, sourcegraph.Term("a")},
			want: &sourcegraph.Plan{
				Repos: &sourcegraph.RepoListOptions{Owner: "u", Query: "a", NoFork: true, Sort: "updated", Direction: "desc"},
				Users: &sourcegraph.UsersListOptions{Query: "u a"},
				Defs: &sourcegraph.DefListOptions{
					Query:    "a",
					RepoRevs: []string{"rr@a"},
					Nonlocal: true,
				},
				TreeRepoRevs: []string{"rr@a"},
				Tree: &sourcegraph.RepoTreeSearchOptions{
					SearchOptions: vcs.SearchOptions{
						Query:     "a",
						QueryType: vcs.FixedQuery,
						N:         10,
					},
				},
			},
			mockReposList: func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
				return &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "rr", DefaultBranch: "a"}}}, nil
			},
		},

		// repo:@u expands to all repos owned by u, and since there
		// are >1, the defs query filters by Exported=true.
		{
			tokens: []sourcegraph.Token{sourcegraph.UserToken{Login: "u"}, sourcegraph.Term("a")},
			want: &sourcegraph.Plan{
				Repos: &sourcegraph.RepoListOptions{Owner: "u", Query: "a", NoFork: true, Sort: "updated", Direction: "desc"},
				Users: &sourcegraph.UsersListOptions{Query: "u a"},
				Defs: &sourcegraph.DefListOptions{
					Query:    "a",
					RepoRevs: []string{"rr0@a", "rr1@b"},
					Nonlocal: true,
					Exported: true,
				},
				TreeRepoRevs: []string{"rr0@a", "rr1@b"},
				Tree: &sourcegraph.RepoTreeSearchOptions{
					SearchOptions: vcs.SearchOptions{
						Query:     "a",
						QueryType: vcs.FixedQuery,
						N:         10,
					},
				},
			},
			mockReposList: func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
				return &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "rr0", DefaultBranch: "a"}, {URI: "rr1", DefaultBranch: "b"}}}, nil
			},
		},
	}
	for _, test := range tests {
		label := "<< " + debugFormatTokens(test.tokens) + " >> "

		ctx := svc.WithServices(context.Background(), svc.Services{
			Repos: &mock.ReposServer{List_: test.mockReposList},
		})

		plan, err := NewPlan(ctx, test.tokens)
		if !reflect.DeepEqual(err, test.wantErr) {
			if test.wantErr == nil {
				t.Errorf("%s: NewPlan: %s", label, err)
			} else {
				t.Errorf("%s: NewPlan: got error %q, want %q", label, err, test.wantErr)
			}
			return
		}
		if err != nil {
			return
		}
		if !reflect.DeepEqual(plan, test.want) {
			t.Errorf("%s: got plan\n%s\n\nwant\n%s", label, plan, test.want)
		}
	}
}
