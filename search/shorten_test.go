package search

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph/mock"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/svc"
)

func TestShorten(t *testing.T) {
	tests := []struct {
		tokens  []sourcegraph.Token
		want    []string
		wantErr error

		currentUserLogin string
		mockOrgsList     func(ctx context.Context, op *sourcegraph.OrgsListOp) (*sourcegraph.OrgList, error)
	}{
		{
			tokens: []sourcegraph.Token{
				sourcegraph.RepoToken{URI: "github.com/o/repo"},
				sourcegraph.Term("t"),
			},
			want:             []string{"repo", "t"},
			currentUserLogin: "u",
			mockOrgsList: func(ctx context.Context, op *sourcegraph.OrgsListOp) (*sourcegraph.OrgList, error) {
				return &sourcegraph.OrgList{
					Orgs: []*sourcegraph.Org{{User: sourcegraph.User{Login: "o"}}},
				}, nil
			},
		},
		{
			tokens: []sourcegraph.Token{
				sourcegraph.RepoToken{URI: "github.com/u/repo"},
				sourcegraph.Term("t"),
			},
			want:             []string{"repo", "t"},
			currentUserLogin: "u",
			mockOrgsList: func(ctx context.Context, op *sourcegraph.OrgsListOp) (*sourcegraph.OrgList, error) {
				return &sourcegraph.OrgList{
					Orgs: []*sourcegraph.Org{{User: sourcegraph.User{Login: "o"}}},
				}, nil
			},
		},
	}
	for _, test := range tests {
		label := "<< " + debugFormatTokens(test.tokens) + " >> "

		ctx := svc.WithServices(context.Background(), svc.Services{
			Orgs: &mock.OrgsServer{
				List_: test.mockOrgsList,
			},
		})
		ctx = auth.WithActor(ctx, auth.Actor{UID: 1, Login: test.currentUserLogin})

		shortened, err := Shorten(ctx, test.tokens)
		if !reflect.DeepEqual(err, test.wantErr) {
			if test.wantErr == nil {
				t.Errorf("%s: Shorten: %s", label, err)
			} else {
				t.Errorf("%s: Shorten: got error %q, want %q", label, err, test.wantErr)
			}
			continue
		}
		if err != nil {
			continue
		}

		if !reflect.DeepEqual(shortened, test.want) {
			t.Errorf("%s: got shortened tokens\n%s\n\nwant\n%s", label, asJSON(shortened), asJSON(test.want))
		}
	}
}
