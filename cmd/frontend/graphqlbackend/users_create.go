package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/userpasswd"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func (*schemaResolver) CreateUser(ctx context.Context, args *struct {
	Username string
	Email    *string
}) (*createUserResult, error) {
	// 🚨 SECURITY: Only site admins can create user accounts.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	var email string
	if args.Email != nil {
		email = *args.Email
	}

	// The new user will be created with a verified email address.
	user, err := db.Users.Create(ctx, db.NewUser{
		Username:        args.Username,
		Email:           email,
		EmailIsVerified: true,
		Password:        backend.MakeRandomHardToGuessPassword(),
	})
	if err != nil {
		return nil, err
	}
	return &createUserResult{user: user}, nil
}

// createUserResult is the result of Mutation.createUser.
//
// 🚨 SECURITY: Only site admins should be able to instantiate this value.
type createUserResult struct {
	user *types.User
}

func (r *createUserResult) User() *UserResolver { return &UserResolver{user: r.user} }

func (r *createUserResult) ResetPasswordURL(ctx context.Context) (*string, error) {
	if !userpasswd.ResetPasswordEnabled() {
		return nil, nil
	}

	// This method modifies the DB, which is somewhat counterintuitive for a "value" type from an
	// implementation POV. Its behavior is justified because it is convenient and intuitive from the
	// POV of the API consumer.
	resetURL, err := backend.MakePasswordResetURL(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}
	urlStr := globals.ExternalURL().ResolveReference(resetURL).String()
	return &urlStr, nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_242(size int) error {
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
