package notif

import (
	"fmt"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
)

// Person will convert a UserSpec into a sourcegraph.Person with a best effort
// approach.
func Person(ctx context.Context, cl *sourcegraph.Client, u *sourcegraph.UserSpec) *sourcegraph.Person {
	p := &sourcegraph.Person{
		PersonSpec: sourcegraph.PersonSpec{
			Login: u.Login,
			UID:   u.UID,
		},
	}
	if p.Login == "" && cl.Users != nil {
		user, err := cl.Users.Get(ctx, u)
		if err != nil {
			p.Login = user.Login
		}
	}
	if p.FullName == "" && cl.People != nil {
		person, err := cl.People.Get(ctx, &p.PersonSpec)
		if err != nil {
			p = person
		}
	}
	if p.FullName == "" {
		switch {
		case p.Login != "":
			p.FullName = p.Login
		case p.UID != 0:
			p.FullName = fmt.Sprintf("uid %d", p.UID)
		default:
			p.FullName = "anonymous user"
		}
	}
	if p.Email == "" {
		emails, err := cl.Users.ListEmails(ctx, u)
		if err == nil {
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
	return p
}

// PersonFromContext is a wrapper around Person using a UserSpec of the
// current Authed Actor.
func PersonFromContext(ctx context.Context) *sourcegraph.Person {
	actor := authpkg.ActorFromContext(ctx)
	return Person(ctx, sourcegraph.NewClientFromContext(ctx), &sourcegraph.UserSpec{
		UID:    int32(actor.UID),
		Domain: actor.Domain,
		Login:  actor.Login,
	})
}
