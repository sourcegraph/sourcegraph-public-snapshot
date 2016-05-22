package backend

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"

	"strings"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
	"sourcegraph.com/sourcegraph/sourcegraph/util/emailaddrs"
)

var People sourcegraph.PeopleServer = &people{}

type people struct{}

var _ sourcegraph.PeopleServer = (*people)(nil)

func (s *people) Get(ctx context.Context, personSpec *sourcegraph.PersonSpec) (*sourcegraph.Person, error) {
	var (
		userSpec *sourcegraph.UserSpec
		p        *sourcegraph.Person
	)

	// If only email is set, look up in directory service.
	if personSpec.UID == 0 && personSpec.Login == "" && personSpec.Email != "" {
		var err error
		userSpec, err = store.DirectoryFromContext(ctx).GetUserByEmail(ctx, personSpec.Email)
		_, isNotExist := err.(*store.UserNotFoundError)
		if userSpec == nil || isNotExist {
			p = newTransientPerson(personSpec.Email)
		} else if err != nil {
			return nil, err
		}
	} else {
		userSpec = &sourcegraph.UserSpec{UID: personSpec.UID, Login: personSpec.Login}
	}

	if p == nil {
		u, err := svc.Users(ctx).Get(ctx, userSpec)
		if err != nil {
			return nil, err
		}
		p = u.Person()

		// Fallback on gravatarURL if avatarURL is not available
		if p.AvatarURL == "" {
			p.AvatarURL = gravatarURL(p.PersonSpec.Email, 0)
		}
	}

	return p, nil
}

// newTransientPerson returns a new user struct for a
// transient user (see the Transient field documentation on the User struct for
// more information).
func newTransientPerson(email string) *sourcegraph.Person {
	oe, err := emailaddrs.Obfuscate(email)
	if err != nil {
		log.Printf("Error obfuscating email %q: %s", email, err)
		oe, _ = emailaddrs.Obfuscate("error@example.com")
	}

	return &sourcegraph.Person{
		PersonSpec: sourcegraph.PersonSpec{Email: oe},
		AvatarURL:  gravatarURL(email, 0),
	}
}

// gravatarURL returns the URL to the Gravatar avatar image for
// email. If size is 0, the default is used.
func gravatarURL(email string, size uint16) string {
	if size == 0 {
		size = 128
	}
	email = strings.TrimSpace(email) // Trim leading and trailing whitespace from an email address.
	email = strings.ToLower(email)   // Force all characters to lower-case.
	h := md5.New()
	io.WriteString(h, email) // md5 hash the final string.
	return fmt.Sprintf("https://secure.gravatar.com/avatar/%x?s=%d&d=mm", h.Sum(nil), size)
}
