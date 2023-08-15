package commit

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestQueryToGitQuery(t *testing.T) {
	type testCase struct {
		name   string
		input  query.Basic
		diff   bool
		output protocol.Node
	}

	cases := []testCase{{
		name: "negated repo does not result in nil node (#26032)",
		input: query.Basic{
			Parameters: []query.Parameter{{Field: query.FieldRepo, Negated: true}},
		},
		diff:   false,
		output: &protocol.Boolean{Value: true},
	}, {
		name: "expensive nodes are placed last",
		input: query.Basic{
			Parameters: []query.Parameter{{Field: query.FieldAuthor, Value: "b"}},
			Pattern:    query.Pattern{Value: "a"},
		},
		diff: true,
		output: protocol.NewAnd(
			&protocol.AuthorMatches{Expr: "b", IgnoreCase: true},
			&protocol.DiffMatches{Expr: "a", IgnoreCase: true},
		),
	}, {
		name: "all supported nodes are converted",
		input: query.Basic{
			Parameters: []query.Parameter{
				{Field: query.FieldAuthor, Value: "author"},
				{Field: query.FieldCommitter, Value: "committer"},
				{Field: query.FieldBefore, Value: "2021-09-10"},
				{Field: query.FieldAfter, Value: "2021-09-08"},
				{Field: query.FieldFile, Value: "file"},
				{Field: query.FieldMessage, Value: "message1"},
			},
			Pattern: query.Pattern{Value: "message2"},
		},
		diff: false,
		output: protocol.NewAnd(
			&protocol.CommitBefore{Time: time.Date(2021, 9, 10, 0, 0, 0, 0, time.UTC)},
			&protocol.CommitAfter{Time: time.Date(2021, 9, 8, 0, 0, 0, 0, time.UTC)},
			&protocol.AuthorMatches{Expr: "author", IgnoreCase: true},
			&protocol.CommitterMatches{Expr: "committer", IgnoreCase: true},
			&protocol.MessageMatches{Expr: "message1", IgnoreCase: true},
			&protocol.MessageMatches{Expr: "message2", IgnoreCase: true},
			&protocol.DiffModifiesFile{Expr: "file", IgnoreCase: true},
		),
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := QueryToGitQuery(tc.input, tc.diff)
			require.Equal(t, tc.output, output)
		})
	}
}

func TestExpandUsernamesToEmails(t *testing.T) {
	users := dbmocks.NewStrictMockUserStore()
	users.GetByUsernameFunc.SetDefaultHook(func(_ context.Context, username string) (*types.User, error) {
		if want := "alice"; username != want {
			t.Errorf("got %q, want %q", username, want)
		}
		return &types.User{ID: 123}, nil
	})

	userEmails := dbmocks.NewStrictMockUserEmailsStore()
	userEmails.ListByUserFunc.SetDefaultHook(func(_ context.Context, opt database.UserEmailsListOptions) ([]*database.UserEmail, error) {
		if want := int32(123); opt.UserID != want {
			t.Errorf("got %v, want %v", opt.UserID, want)
		}
		t := time.Now()
		return []*database.UserEmail{
			{Email: "alice@example.com", VerifiedAt: &t},
			{Email: "alice@example.org", VerifiedAt: &t},
		}, nil
	})

	db := dbmocks.NewStrictMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.UserEmailsFunc.SetDefaultReturn(userEmails)

	x, err := expandUsernamesToEmails(context.Background(), db, []string{"foo", "@alice"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"foo", `alice@example\.com`, `alice@example\.org`}; !reflect.DeepEqual(x, want) {
		t.Errorf("got %q, want %q", x, want)
	}
}
