package fs

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/rwvfs"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/randstring"
)

// accounts is a FS-backed implementation of the Accounts store.
type accounts struct{}

var _ store.Accounts = (*accounts)(nil)

func (s *accounts) GetByGitHubID(ctx context.Context, id int) (*sourcegraph.User, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "GetByGitHubID")
}

func (s *accounts) Create(ctx context.Context, newUser *sourcegraph.User) (*sourcegraph.User, error) {
	if newUser.UID != 0 && !authutil.ActiveFlags.MigrateMode {
		return nil, errors.New("uid already set")
	}
	if newUser.Login == "" {
		return nil, errors.New("login must be set")
	}

	users, err := readUserDB(ctx)
	if err != nil {
		return nil, err
	}

	maxUID := int32(0)

	// Verify login and UID uniqueness.
	for _, user := range users {
		if user.UID > maxUID {
			maxUID = user.UID
		}
		if user.Login == newUser.Login || user.UID == newUser.UID {
			return nil, &store.AccountAlreadyExistsError{Login: newUser.Login, UID: newUser.UID}
		}
	}

	if newUser.UID == 0 {
		newUser.UID = maxUID + 1
	}
	users = append(users, &userDBEntry{User: *newUser})

	if err := writeUserDB(ctx, users); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *accounts) Update(ctx context.Context, modUser *sourcegraph.User) error {
	users, err := readUserDB(ctx)
	if err != nil {
		return err
	}

	for i, user := range users {
		if user.UID == modUser.UID {
			users[i].User = *modUser
			return writeUserDB(ctx, users)
		}
	}

	return &store.UserNotFoundError{UID: int(modUser.UID)}
}

func (s *accounts) UpdateEmails(ctx context.Context, user sourcegraph.UserSpec, emails []*sourcegraph.EmailAddr) error {
	users, err := readUserDB(ctx)
	if err != nil {
		return err
	}

	for i, u := range users {
		if u.UID == user.UID {
			users[i].EmailAddrs = emails
			return writeUserDB(ctx, users)
		}
	}

	return &store.UserNotFoundError{UID: int(user.UID)}
}

func (s *accounts) Delete(ctx context.Context, uid int32) error {
	users, err := readUserDB(ctx)
	if err != nil {
		return err
	}

	for i, u := range users {
		if u.UID == uid {
			users[i] = users[len(users)-1]
			users = users[:len(users)-1]
			return writeUserDB(ctx, users)
		}
	}

	return &store.UserNotFoundError{UID: int(uid)}
}

const passwordResetFilename = "password_reset.json"

type passwordReset struct {
	Token string
	UID   int32
}

// readPasswordResetDB reads the password reset requests db from disk.
// If no such file exists, an empty slice is returned (and no error).
func readPasswordResetDB(ctx context.Context) ([]*passwordReset, error) {
	f, err := dbVFS(ctx).Open(passwordResetFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var requests []*passwordReset
	if err := json.NewDecoder(f).Decode(&requests); err != nil {
		return nil, err
	}
	return requests, nil
}

// writePasswordResetDB writes the password reset requests db to disk.
func writePasswordResetDB(ctx context.Context, users []*passwordReset) (err error) {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	if err := rwvfs.MkdirAll(dbVFS(ctx), "."); err != nil {
		return err
	}
	f, err := dbVFS(ctx).Create(passwordResetFilename)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := f.Close(); err2 != nil {
			if err == nil {
				err = err2
			} else {
				log.Printf("Warning: closing password reset DB after error (%s) failed: %s.", err, err2)
			}
		}
	}()

	_, err = f.Write(data)
	return err
}

func (s *accounts) RequestPasswordReset(ctx context.Context, user *sourcegraph.User) (*sourcegraph.PasswordResetToken, error) {
	const tokenLength = 44
	if user.UID == 0 {
		return nil, errors.New("UID must be set")
	}

	requests, err := readPasswordResetDB(ctx)
	if err != nil {
		return nil, err
	}

	token := randstring.NewLen(tokenLength)
	requests = append(requests, &passwordReset{
		Token: token,
		UID:   user.UID,
	})

	if err := writePasswordResetDB(ctx, requests); err != nil {
		return nil, err
	}
	return &sourcegraph.PasswordResetToken{Token: token}, nil
}

func (s *accounts) ResetPassword(ctx context.Context, newPass *sourcegraph.NewPassword) error {
	genericErr := errors.New("error reseting password") // don't need to reveal everything
	requests, err := readPasswordResetDB(ctx)
	if err != nil {
		return err
	}

	for i := range requests {
		if subtle.ConstantTimeCompare([]byte(newPass.Token.Token), []byte(requests[i].Token)) == 1 {
			log15.Info("Resetting password", "store", "Accounts", "UID", requests[i].UID)
			if err := (password{}).SetPassword(ctx, requests[i].UID, newPass.Password); err != nil {
				return fmt.Errorf("Error changing password: %s", err)
			}

			requests[i] = requests[len(requests)-1]
			requests = requests[:len(requests)-1]

			// Save to disk
			if err := writePasswordResetDB(ctx, requests); err != nil {
				return err
			}
			return nil
		}
	}

	log15.Warn("Token does not exist in password reset database", "store", "Accounts")
	return genericErr
}
