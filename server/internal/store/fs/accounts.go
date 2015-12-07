package fs

import (
	"errors"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// Accounts is a FS-backed implementation of the Accounts store.
type Accounts struct{}

var _ store.Accounts = (*Accounts)(nil)

func (s *Accounts) GetByGitHubID(ctx context.Context, id int) (*sourcegraph.User, error) {
	return nil, &sourcegraph.NotImplementedError{What: "GetByGitHubID"}
}

func (s *Accounts) Create(ctx context.Context, newUser *sourcegraph.User) (*sourcegraph.User, error) {
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

func (s *Accounts) Update(ctx context.Context, modUser *sourcegraph.User) error {
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

func (s *Accounts) UpdateEmails(ctx context.Context, user sourcegraph.UserSpec, emails []*sourcegraph.EmailAddr) error {
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

func (s *Accounts) RequestPasswordReset(ctx context.Context, uid *sourcegraph.User) (*sourcegraph.PasswordResetToken, error) {
	return nil, &sourcegraph.NotImplementedError{What: "file system user password reset"}
}

func (s *Accounts) ResetPassword(ctx context.Context, newPass *sourcegraph.NewPassword) error {
	return &sourcegraph.NotImplementedError{What: "file system user password reset"}
}
