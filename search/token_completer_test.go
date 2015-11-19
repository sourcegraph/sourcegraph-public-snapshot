package search

import (
	"encoding/json"
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph/mock"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/svc"
)

func TestCompleteToken(t *testing.T) {
	tests := []struct {
		partial sourcegraph.Token
		scope   []sourcegraph.Token
		want    []sourcegraph.Token
		wantErr error

		mockReposList              func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error)
		mockReposGet               func(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (*sourcegraph.Repo, error)
		mockBuildsGetRepoBuildInfo func(ctx context.Context, op *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error)
		mockUsersList              func(ctx context.Context, opt *sourcegraph.UsersListOptions) (*sourcegraph.UserList, error)
		mockOrgsList               func(ctx context.Context, op *sourcegraph.OrgsListOp) (*sourcegraph.OrgList, error)
		mockDefsList               func(ctx context.Context, opt *sourcegraph.DefListOptions) (*sourcegraph.DefList, error)
		mockUnitsGet               func(ctx context.Context, unitSpec *sourcegraph.UnitSpec) (*unit.RepoSourceUnit, error)
		mockUnitsList              func(ctx context.Context, opt *sourcegraph.UnitListOptions) (*sourcegraph.RepoSourceUnitList, error)
	}{
		// Prefetch
		{
			scope: []sourcegraph.Token{},
			want: []sourcegraph.Token{
				sourcegraph.RepoToken{URI: "r", Repo: &sourcegraph.Repo{URI: "r"}},
				sourcegraph.UserToken{Login: "o", User: &sourcegraph.User{Login: "o", IsOrganization: true}},
				sourcegraph.UserToken{Login: "u", User: &sourcegraph.User{Login: "u"}},
			},
			mockReposList: func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
				return &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "r"}}}, nil
			},
			mockUsersList: func(ctx context.Context, opt *sourcegraph.UsersListOptions) (*sourcegraph.UserList, error) {
				return &sourcegraph.UserList{Users: []*sourcegraph.User{{Login: "u"}}}, nil
			},
			mockOrgsList: func(ctx context.Context, op *sourcegraph.OrgsListOp) (*sourcegraph.OrgList, error) {
				return &sourcegraph.OrgList{Orgs: []*sourcegraph.Org{{User: sourcegraph.User{Login: "o", IsOrganization: true}}}}, nil
			},
		},

		// Completion
		{
			partial: sourcegraph.AnyToken("a"),
			want: []sourcegraph.Token{
				sourcegraph.RepoToken{URI: "r", Repo: &sourcegraph.Repo{URI: "r"}},
				sourcegraph.UserToken{Login: "u", User: &sourcegraph.User{Login: "u"}},
			},
			mockReposList: func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
				return &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "r"}}}, nil
			},
			mockUsersList: func(ctx context.Context, opt *sourcegraph.UsersListOptions) (*sourcegraph.UserList, error) {
				return &sourcegraph.UserList{Users: []*sourcegraph.User{{Login: "u"}}}, nil
			},
			mockOrgsList: func(ctx context.Context, op *sourcegraph.OrgsListOp) (*sourcegraph.OrgList, error) {
				return &sourcegraph.OrgList{Orgs: []*sourcegraph.Org{{User: sourcegraph.User{Login: "o", IsOrganization: true}}}}, nil
			},
		},
		{
			partial: sourcegraph.RepoToken{URI: "r"},
			want:    []sourcegraph.Token{sourcegraph.RepoToken{URI: "rr", Repo: &sourcegraph.Repo{URI: "rr"}}},
			mockReposList: func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
				return &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "rr"}}}, nil
			},
		},
		{
			partial: sourcegraph.AnyToken("a"),
			scope:   []sourcegraph.Token{sourcegraph.RepoToken{URI: "r"}},
			want:    []sourcegraph.Token{sourcegraph.Term("abc")},
			mockReposList: func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
				return &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "rr"}}}, nil
			},
			mockDefsList: func(ctx context.Context, opt *sourcegraph.DefListOptions) (*sourcegraph.DefList, error) {
				return &sourcegraph.DefList{Defs: []*sourcegraph.Def{{Def: graph.Def{Name: "abc"}}}}, nil
			},
			mockReposGet: func(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
				return &sourcegraph.Repo{URI: "rr"}, nil
			},
		},
		{
			partial: sourcegraph.UnitToken{Name: "u"},
			scope:   []sourcegraph.Token{sourcegraph.RepoToken{URI: "rr", Repo: &sourcegraph.Repo{URI: "rr"}}},
			want: []sourcegraph.Token{
				sourcegraph.UnitToken{Name: "uu", UnitType: "t", Unit: &unit.RepoSourceUnit{Unit: "uu", UnitType: "t"}},
			},
			mockUnitsList: func(ctx context.Context, opt *sourcegraph.UnitListOptions) (*sourcegraph.RepoSourceUnitList, error) {
				return &sourcegraph.RepoSourceUnitList{Units: []*unit.RepoSourceUnit{{Unit: "uu", UnitType: "t"}}}, nil
			},
			mockReposGet: func(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
				return &sourcegraph.Repo{URI: "rr"}, nil
			},
			mockBuildsGetRepoBuildInfo: reposGetBuildOK,
		},
		{
			partial: sourcegraph.AnyToken("uuu"),
			scope:   []sourcegraph.Token{sourcegraph.RepoToken{URI: "rr", Repo: &sourcegraph.Repo{URI: "rr"}}},
			want: []sourcegraph.Token{
				sourcegraph.UnitToken{Name: "uuuu", UnitType: "t", Unit: &unit.RepoSourceUnit{Unit: "uuuu", UnitType: "t"}},
			},
			mockUnitsList: func(ctx context.Context, opt *sourcegraph.UnitListOptions) (*sourcegraph.RepoSourceUnitList, error) {
				return &sourcegraph.RepoSourceUnitList{Units: []*unit.RepoSourceUnit{{Unit: "uuuu", UnitType: "t"}}}, nil
			},
			mockDefsList: func(ctx context.Context, opt *sourcegraph.DefListOptions) (*sourcegraph.DefList, error) {
				return &sourcegraph.DefList{}, nil
			},
			mockReposList: func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
				return &sourcegraph.RepoList{}, nil
			},
			mockReposGet: func(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
				return &sourcegraph.Repo{URI: "rr"}, nil
			},
			mockBuildsGetRepoBuildInfo: reposGetBuildOK,
			mockUsersList: func(ctx context.Context, opt *sourcegraph.UsersListOptions) (*sourcegraph.UserList, error) {
				return &sourcegraph.UserList{}, nil
			},
			mockOrgsList: func(ctx context.Context, op *sourcegraph.OrgsListOp) (*sourcegraph.OrgList, error) {
				return &sourcegraph.OrgList{}, nil
			},
		},
		{
			partial: sourcegraph.UserToken{Login: "u"},
			want:    []sourcegraph.Token{sourcegraph.UserToken{Login: "uu", User: &sourcegraph.User{Login: "uu"}}},
			mockUsersList: func(ctx context.Context, opt *sourcegraph.UsersListOptions) (*sourcegraph.UserList, error) {
				return &sourcegraph.UserList{Users: []*sourcegraph.User{{Login: "uu"}}}, nil
			},
		},
	}
	for _, test := range tests {
		label := "<< " + debugFormatTokens([]sourcegraph.Token{test.partial}) + " || " + debugFormatTokens(test.scope) + " >> "

		ctx := svc.WithServices(context.Background(), svc.Services{
			Repos: &mock.ReposServer{
				Get_:  test.mockReposGet,
				List_: test.mockReposList,
			},
			Builds: &mock.BuildsServer{
				GetRepoBuildInfo_: test.mockBuildsGetRepoBuildInfo,
			},
			Units: &mock.UnitsServer{
				Get_:  test.mockUnitsGet,
				List_: test.mockUnitsList,
			},
			Users: &mock.UsersServer{
				List_: test.mockUsersList,
			},
			Orgs: &mock.OrgsServer{
				List_: test.mockOrgsList,
			},
			Defs: &mock.DefsServer{List_: test.mockDefsList},
		})
		ctx = auth.WithActor(ctx, auth.Actor{UID: 1, Login: "u"})

		conf := TokenCompletionConfig{DontResolveDefs: test.mockDefsList == nil}
		comps, err := CompleteToken(ctx, test.partial, test.scope, conf)
		if !reflect.DeepEqual(err, test.wantErr) {
			if test.wantErr == nil {
				t.Errorf("%s: CompleteToken: %s", label, err)
			} else {
				t.Errorf("%s: CompleteToken: got error %q, want %q", label, err, test.wantErr)
			}
			continue
		}
		if err != nil {
			continue
		}
		if !reflect.DeepEqual(comps, test.want) {
			t.Errorf("%s: got completions\n%s\n\nwant\n%s", label, debugFormatTokens(comps), debugFormatTokens(test.want))
		}
	}
}

func asJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}
