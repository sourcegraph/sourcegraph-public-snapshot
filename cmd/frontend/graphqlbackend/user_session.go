package graphqlbackend

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func (r *UserResolver) Session(ctx context.Context) (*sessionResolver, error) {
	// 🚨 SECURITY: Only the user can view their session information, because it is retrieved from
	// the context of this request (and not persisted in a way that is queryable).
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() || actor.UID != r.user.ID {
		return nil, errors.New("unable to view session for a user other than the currently authenticated user")
	}

	var sr sessionResolver
	if actor.FromSessionCookie {
		// The http-header auth provider is the only auth provider that a user can't sign out from.
		for _, p := range conf.Get().Critical.AuthProviders {
			if p.HttpHeader == nil {
				sr.canSignOut = true
				break
			}
		}
	}
	return &sr, nil
}

type sessionResolver struct {
	canSignOut bool
}

func (r *sessionResolver) CanSignOut() bool { return r.canSignOut }

// random will create a file of size bytes (rounded up to next 1024 size)
func random_237(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
