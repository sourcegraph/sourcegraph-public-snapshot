package mdutil

import (
	"regexp"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// mentionsPattern is a regular expression that matches user and email mentions of
// the form @user or @user@domain.
var mentionsPattern = regexp.MustCompile("\\B@[@.a-zA-Z0-9_-]+")

// Mentions returns the list of people (users or emails) mentioned in the passed
// argument.
func Mentions(ctx context.Context, md []byte) ([]*sourcegraph.UserSpec, error) {
	indexes := mentionsPattern.FindAllIndex(md, -1)
	if len(indexes) == 0 {
		return []*sourcegraph.UserSpec{}, nil
	}
	ppl := make([]*sourcegraph.UserSpec, 0, len(indexes))
	cl := sourcegraph.NewClientFromContext(ctx)
	for _, idx := range indexes {
		m := md[idx[0]+1 : idx[1]]
		u, err := findPerson(ctx, cl, m)
		if err != nil {
			if grpc.Code(err) == codes.NotFound {
				continue
			}
			return nil, err
		}
		ppl = append(ppl, u)
	}
	return ppl, nil
}

// findPerson attempts to resolve the passed mention as an existing person or as
// a valid email.
func findPerson(ctx context.Context, cl *sourcegraph.Client, mention []byte) (*sourcegraph.UserSpec, error) {
	m := string(mention)
	// is this a person?
	u, err := cl.Users.Get(ctx, &sourcegraph.UserSpec{Login: m})
	if err != nil {
		return nil, err
	}
	s := u.Spec()
	return &s, nil
}
