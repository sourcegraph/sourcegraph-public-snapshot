package backend

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/randstring"
)

func MakeRandomHardToGuessPassword() string {
	return randstring.NewLen(36)
}

func MakePasswordResetURL(ctx context.Context, userID int32) (*url.URL, error) {
	resetCode, err := db.Users.RenewPasswordResetCode(ctx, userID)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("userID", strconv.Itoa(int(userID)))
	query.Set("code", resetCode)
	return &url.URL{Path: "/password-reset", RawQuery: query.Encode()}, nil
}

// CheckActorHasTag reports whether the context actor has the given tag. If not, or if an error
// occurs, a non-nil error is returned.
func CheckActorHasTag(ctx context.Context, tag string) error {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return ErrNotAuthenticated
	}
	user, err := db.Users.GetByID(ctx, actor.UID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrNotAuthenticated
	}
	for _, t := range user.Tags {
		if t == tag {
			return nil
		}
	}
	return fmt.Errorf("actor lacks required tag %q", tag)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_33(size int) error {
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
