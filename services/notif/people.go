package notif

import (
	"fmt"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
)

// Person will convert a UserSpec into a sourcegraph.Person with a best effort
// approach.
func Person(ctx context.Context, u *sourcegraph.UserSpec) *sourcegraph.Person {
	if u.UID == 0 && u.Login == "" {
		return nil
	}

	p := &sourcegraph.Person{
		PersonSpec: sourcegraph.PersonSpec{
			Login: u.Login,
			UID:   u.UID,
		},
	}

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		log15.Error("notif.Person", "error", err)
		return p
	}

	if p.Login == "" && cl.Users != nil {
		user, err := cl.Users.Get(ctx, u)
		if err != nil {
			log15.Warn("notif.Person", "ignoring", err)
		} else {
			p.Login = user.Login
		}
	}
	if p.FullName == "" && cl.People != nil {
		person, err := cl.People.Get(ctx, &p.PersonSpec)
		if err != nil {
			log15.Warn("notif.Person", "ignoring", err)
		} else {
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
		// An error will occur when one user tries to retrieve the email
		// of a different user since ListEmails will only return emails that
		// the user has permission to view.
		if err != nil {
			log15.Warn("notif.Person", "ignoring", err)
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

	return p
}

// PersonFromContext is a wrapper around Person using a UserSpec of the
// current Authed Actor.
func PersonFromContext(ctx context.Context) *sourcegraph.Person {
	return Person(ctx, UserFromContext(ctx))
}

// UserFromContext will create a UserSpec based on the current Authed Actor
func UserFromContext(ctx context.Context) *sourcegraph.UserSpec {
	actor := authpkg.ActorFromContext(ctx)
	return &sourcegraph.UserSpec{
		UID:   int32(actor.UID),
		Login: actor.Login,
	}
}
