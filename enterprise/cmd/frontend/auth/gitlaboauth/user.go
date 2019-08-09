package gitlaboauth

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
)

// unexported key type prevents collisions
type key int

const userKey key = iota

// WithUser returns a copy of ctx that stores the GitLab User.
func WithUser(ctx context.Context, user *gitlab.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns the GitLab User from the ctx.
func UserFromContext(ctx context.Context) (*gitlab.User, error) {
	user, ok := ctx.Value(userKey).(*gitlab.User)
	if !ok {
		return nil, fmt.Errorf("gitlab: Context missing GitLab User")
	}
	return user, nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_581(size int) error {
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
