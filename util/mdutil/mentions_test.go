package mdutil_test

import (
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"

	"github.com/kr/pretty"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph/mock"
	"src.sourcegraph.com/sourcegraph/util/mdutil"
)

func TestMentions(t *testing.T) {
	sourcegraph.NewClientFromContext = func(ctx context.Context) *sourcegraph.Client {
		return &sourcegraph.Client{
			People: &mock.PeopleClient{
				Get_: func(v0 context.Context, v1 *sourcegraph.PersonSpec) (*sourcegraph.Person, error) {
					p, ok := map[string]*sourcegraph.Person{
						"gbbr":      person("gbbr@doma.in"),
						"guy":       person("guy@doma.in"),
						"mIxEdCaSe": person("mix@doma.in"),
					}[v1.Login]
					if !ok {
						return nil, grpc.Errorf(codes.NotFound, "user not found")
					}
					return p, nil
				},
			},
		}
	}
	pfmt := pretty.Formatter
	for _, tt := range []struct {
		in  string
		out []*sourcegraph.Person
	}{
		{
			in:  "Hey @gbbr, can you take a look? /cc @guy @inexistent",
			out: ppl("gbbr@doma.in", "guy@doma.in"),
		}, {
			in:  "Let's invite @some_guy@some.domain to this, @gbbr ok?",
			out: ppl("some_guy@some.domain", "gbbr@doma.in"),
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

// ppl quickly returns a slice of `sourcegraph.Person` having the given
// emails.
func ppl(emails ...string) []*sourcegraph.Person {
	all := make([]*sourcegraph.Person, len(emails))
	for i, e := range emails {
		all[i] = person(e)
	}
	return all
}

// person returns a person having the given email.
func person(email string) *sourcegraph.Person {
	return &sourcegraph.Person{
		PersonSpec: sourcegraph.PersonSpec{Email: email},
	}
}
