package mdutil

import (
	"net/mail"
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
func Mentions(ctx context.Context, md []byte) ([]*sourcegraph.Person, error) {
	indexes := mentionsPattern.FindAllIndex(md, -1)
	if len(indexes) == 0 {
		return []*sourcegraph.Person{}, nil
	}
	ppl := make([]*sourcegraph.Person, 0, len(indexes))
	pplsvc := sourcegraph.NewClientFromContext(ctx).People
	for _, idx := range indexes {
		m := md[idx[0]+1 : idx[1]]
		p, err := findPerson(ctx, pplsvc, m)
		if err != nil {
			if grpc.Code(err) == codes.NotFound {
				continue
			}
			return nil, err
		}
		ppl = append(ppl, p)
	}
	return ppl, nil
}

// findPerson attempts to resolve the passed mention as an existing person or as
// a valid email.
func findPerson(ctx context.Context, ppl sourcegraph.PeopleClient, mention []byte) (*sourcegraph.Person, error) {
	m := string(mention)
	// is this an email address?
	if _, err := mail.ParseAddress(m); err == nil {
		return &sourcegraph.Person{
			PersonSpec: sourcegraph.PersonSpec{Email: m},
		}, nil
	}
	// is this a person?
	p, err := ppl.Get(ctx, &sourcegraph.PersonSpec{Login: m})
	if err != nil {
		return nil, err
	}
	return p, nil
}
