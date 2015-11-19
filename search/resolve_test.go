package search

import (
	"errors"
	"reflect"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph/mock"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/srclib/unit"
	"src.sourcegraph.com/sourcegraph/svc"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		tokens          []sourcegraph.Token
		wantValid       bool
		wantResolved    []sourcegraph.Token
		wantResolveErrs []sourcegraph.TokenError
		wantErr         error

		mockReposList              func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error)
		mockReposGet               func(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (*sourcegraph.Repo, error)
		mockReposGetCommit         func(ctx context.Context, repoRevSpec *sourcegraph.RepoRevSpec) (*vcs.Commit, error)
		mockBuildsGetRepoBuildInfo func(ctx context.Context, op *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error)
		mockUsersGet               func(ctx context.Context, userSpec *sourcegraph.UserSpec) (*sourcegraph.User, error)
		mockUsersList              func(ctx context.Context, opt *sourcegraph.UsersListOptions) (*sourcegraph.UserList, error)
		mockDefsGet                func(ctx context.Context, op *sourcegraph.DefsGetOp) (*sourcegraph.Def, error)
		mockDefsList               func(ctx context.Context, opt *sourcegraph.DefListOptions) (*sourcegraph.DefList, error)
		mockUnitsGet               func(ctx context.Context, unitSpec *sourcegraph.UnitSpec) (*unit.RepoSourceUnit, error)
		mockUnitsList              func(ctx context.Context, opt *sourcegraph.UnitListOptions) (*sourcegraph.RepoSourceUnitList, error)
	}{
		{
			tokens:          []sourcegraph.Token{},
			wantValid:       false,
			wantResolveErrs: nil,
		},

		{
			tokens:    []sourcegraph.Token{sourcegraph.AnyToken("r")},
			wantValid: true,
			wantResolved: []sourcegraph.Token{
				sourcegraph.RepoToken{
					URI:  "r",
					Repo: &sourcegraph.Repo{URI: "r"},
				},
			},
			mockReposGet: func(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
				return &sourcegraph.Repo{URI: "r"}, nil
			},
			mockUsersList: func(ctx context.Context, opt *sourcegraph.UsersListOptions) (*sourcegraph.UserList, error) {
				return &sourcegraph.UserList{}, nil
			},
		},
		{
			tokens:    []sourcegraph.Token{sourcegraph.AnyToken("r")},
			wantValid: true,
			wantResolved: []sourcegraph.Token{
				sourcegraph.RepoToken{
					URI:  "rr",
					Repo: &sourcegraph.Repo{URI: "rr"},
				},
			},
			mockReposGet: func(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
				return nil, grpc.Errorf(codes.NotFound, "")
			},
			mockReposList: func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
				return &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "rr"}}}, nil
			},
			mockUsersList: func(ctx context.Context, opt *sourcegraph.UsersListOptions) (*sourcegraph.UserList, error) {
				return &sourcegraph.UserList{}, nil
			},
		},
		{
			tokens:    []sourcegraph.Token{sourcegraph.RepoToken{URI: "r"}},
			wantValid: false,
			wantResolveErrs: []sourcegraph.TokenError{
				sourcegraph.TokenError{
					Index:   1,
					Token:   tp(sourcegraph.RepoToken{URI: "r"}),
					Message: "Repository not found: r",
				},
			},
			mockReposGet: reposGetNone,
		},
		{
			tokens: []sourcegraph.Token{
				sourcegraph.RepoToken{URI: "r"},
				sourcegraph.UnitToken{UnitType: "t", Name: "u"},
			},
			wantValid: true,
			wantResolved: []sourcegraph.Token{
				sourcegraph.RepoToken{
					URI:  "r",
					Repo: &sourcegraph.Repo{URI: "r"},
				},
				sourcegraph.UnitToken{
					UnitType: "t",
					Name:     "u",
					Unit:     &unit.RepoSourceUnit{Repo: "r"},
				},
			},
			mockReposGet: func(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
				return &sourcegraph.Repo{URI: "r"}, nil
			},
			mockBuildsGetRepoBuildInfo: reposGetBuildOK,
			mockUnitsGet: func(ctx context.Context, unitSpec *sourcegraph.UnitSpec) (*unit.RepoSourceUnit, error) {
				return &unit.RepoSourceUnit{Repo: "r"}, nil
			},
		},
		{
			tokens:       []sourcegraph.Token{sourcegraph.UserToken{Login: "u"}},
			wantValid:    false,
			wantResolved: []sourcegraph.Token{sourcegraph.UserToken{Login: "u"}},
			wantResolveErrs: []sourcegraph.TokenError{
				sourcegraph.TokenError{
					Index:   1,
					Token:   tp(sourcegraph.UserToken{Login: "u"}),
					Message: "User/org not found: u",
				},
			},
			mockUsersGet: func(ctx context.Context, userSpec *sourcegraph.UserSpec) (*sourcegraph.User, error) {
				return nil, errors.New("x")
			},
		},

		// Try repo URI prefixes.
		{
			tokens: []sourcegraph.Token{sourcegraph.RepoToken{URI: "o/r"}},
			wantResolved: []sourcegraph.Token{
				sourcegraph.RepoToken{
					URI:  "sourcegraph.com/o/r",
					Repo: &sourcegraph.Repo{URI: "sourcegraph.com/o/r"},
				},
			},
			wantValid: true,
			mockReposGet: func(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
				if repoSpec.URI == "sourcegraph.com/o/r" {
					return &sourcegraph.Repo{URI: "sourcegraph.com/o/r"}, nil
				}
				return nil, errors.New("x")
			},
		},
	}
	for _, test := range tests {
		label := "<< " + debugFormatTokens(test.tokens) + " >> "

		ctx := svc.WithServices(context.Background(), svc.Services{
			Repos: &mock.ReposServer{
				Get_:       test.mockReposGet,
				GetCommit_: test.mockReposGetCommit,
				List_:      test.mockReposList,
			},
			Builds: &mock.BuildsServer{
				GetRepoBuildInfo_: test.mockBuildsGetRepoBuildInfo,
			},
			Users: &mock.UsersServer{
				Get_:  test.mockUsersGet,
				List_: test.mockUsersList,
			},
			Defs: &mock.DefsServer{
				Get_:  test.mockDefsGet,
				List_: test.mockDefsList,
			},
			Units: &mock.UnitsServer{
				Get_:  test.mockUnitsGet,
				List_: test.mockUnitsList,
			},
		})

		resolved, resolveErrs, err := Resolve(ctx, test.tokens)
		if !reflect.DeepEqual(err, test.wantErr) {
			if test.wantErr == nil {
				t.Errorf("%s: Resolve: %s", label, err)
			} else {
				t.Errorf("%s: Resolve: got error\n%q\n\nwant\n%q", label, err, test.wantErr)
			}
			continue
		}
		if err != nil {
			continue
		}
		if !reflect.DeepEqual(resolveErrs, test.wantResolveErrs) {
			t.Errorf("%s: got resolve errors %v, want %v", label, resolveErrs, test.wantResolveErrs)
		}
		if !reflect.DeepEqual(resolved, test.wantResolved) {
			t.Errorf("%s: got resolved %v, want %v", label, resolved, test.wantResolved)
		}
	}
}
