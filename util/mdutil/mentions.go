package mdutil

import (
	"net/mail"
	"regexp"

	"gopkg.in/inconshreveable/log15.v2"

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
	cl := sourcegraph.NewClientFromContext(ctx)
	for _, idx := range indexes {
		m := md[idx[0]+1 : idx[1]]
		p, err := findPerson(ctx, cl, m)
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
func findPerson(ctx context.Context, cl *sourcegraph.Client, mention []byte) (*sourcegraph.Person, error) {
	m := string(mention)
	// is this an email address?
	if _, err := mail.ParseAddress(m); err == nil {
		return &sourcegraph.Person{
			PersonSpec: sourcegraph.PersonSpec{Email: m},
		}, nil
	}
	// is this a person?
	p, err := cl.People.Get(ctx, &sourcegraph.PersonSpec{Login: m})
	if err != nil {
		return nil, err
	}
	// We want the mail for the person if we can get it
	if p.Email == "" {
		// TODO This information could potentially be populated
		// directly by the People service. For now we are manually
		// populating to get Mentions working
		emails, err := cl.Users.ListEmails(ctx, &sourcegraph.UserSpec{UID: p.UID})
		if err != nil {
			log15.Warn("Failed to fetch emails for user", "UID", p.UID)
		} else {
			for _, emailAddr := range emails.EmailAddrs {
				if emailAddr.Blacklisted {
					continue
				}
				p.Email = emailAddr.Email
				if emailAddr.Primary {
					break
				}
			}
		}
	}
	return p, nil
}
