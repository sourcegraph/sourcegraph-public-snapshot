package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/userpasswd"
)

type randomizeUserPasswordResult struct {
	userID int32
}

func (r *randomizeUserPasswordResult) ResetPasswordURL(ctx context.Context) (*string, error) {
	if !userpasswd.ResetPasswordEnabled() {
		return nil, nil
	}

	// This method modifies the DB, which is somewhat counterintuitive for a "value" type from an
	// implementation POV. Its behavior is justified because it is convenient and intuitive from the
	// POV of the API consumer.
	resetURL, err := backend.MakePasswordResetURL(ctx, r.userID)
	if err != nil {
		return nil, err
	}
	urlStr := globals.ExternalURL().ResolveReference(resetURL).String()
	return &urlStr, nil
}

func (*schemaResolver) RandomizeUserPassword(ctx context.Context, args *struct {
	User graphql.ID
}) (*randomizeUserPasswordResult, error) {
	// 🚨 SECURITY: Only site admins can randomize user passwords.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	userID, err := UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	if err := db.Users.RandomizePasswordAndClearPasswordResetRateLimit(ctx, userID); err != nil {
		return nil, err
	}

	return &randomizeUserPasswordResult{userID: userID}, nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_243(size int) error {
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
