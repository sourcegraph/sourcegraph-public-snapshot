package mdutil_test

import (
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"

	"github.com/kr/pretty"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph/mock"
	"src.sourcegraph.com/sourcegraph/util/mdutil"
)

func TestMentions(t *testing.T) {
	sourcegraph.MockNewClientFromContext(func(ctx context.Context) *sourcegraph.Client {
		return &sourcegraph.Client{
			Users: &mock.UsersClient{
				Get_: func(v0 context.Context, v1 *sourcegraph.UserSpec) (*sourcegraph.User, error) {
					p, ok := map[string]*sourcegraph.User{
						"gbbr":      &sourcegraph.User{Login: "gbbr@doma.in"},
						"guy":       &sourcegraph.User{Login: "guy@doma.in"},
						"mIxEdCaSe": &sourcegraph.User{Login: "mix@doma.in"},
					}[v1.Login]
					if !ok {
						return nil, grpc.Errorf(codes.NotFound, "user not found")
					}
					return p, nil
				},
			},
		}
	})
	pfmt := pretty.Formatter
	for _, tt := range []struct {
		in  string
		out []*sourcegraph.UserSpec
	}{
		{
			in:  "Hey @gbbr, can you take a look? /cc @guy @inexistent",
			out: ppl("gbbr@doma.in", "guy@doma.in"),
		}, {
			in:  "I don't know @any of @these @people except @guy",
			out: ppl("guy@doma.in"),
		}, {
			in:  "Why does @mIxEdCaSe have such a weird username?",
			out: ppl("mix@doma.in"),
		}, {
			in:  "This username: @MiXeDcAsE should not be found",
			out: ppl(),
		}, {
			in:  "No mentions",
			out: ppl(),
		},
	} {
		ppl, _ := mdutil.Mentions(context.Background(), []byte(tt.in))
		if !reflect.DeepEqual(ppl, tt.out) {
			t.Fatalf("EXPECTED:\n%# v\nGOT:\n%# v", pfmt(ppl), pfmt(tt.out))
		}
	}
}

// ppl quickly returns a slice of `sourcegraph.UserSpec` having the given
// logins.
func ppl(logins ...string) []*sourcegraph.UserSpec {
	all := make([]*sourcegraph.UserSpec, len(logins))
	for i, e := range logins {
		all[i] = person(e)
	}
	return all
}

// person returns a person having the given email.
func person(login string) *sourcegraph.UserSpec {
	return &sourcegraph.UserSpec{Login: login}
}
