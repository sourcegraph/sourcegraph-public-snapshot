package pgsql

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/randstring"
)

// Accounts is a DB-backed implementation of the Accounts store.
type Accounts struct{}

var _ store.Accounts = (*Accounts)(nil)

func (s *Accounts) GetByGitHubID(ctx context.Context, id int) (*sourcegraph.User, error) {
	return nil, &sourcegraph.NotImplementedError{What: "GetByGitHubID"}
}

func (s *Accounts) Create(ctx context.Context, newUser *sourcegraph.User) (*sourcegraph.User, error) {
	if newUser.UID != 0 && !authutil.ActiveFlags.IsLDAP() {
		return nil, errors.New("uid already set")
	}
	if newUser.Login == "" {
		return nil, errors.New("login must be set")
	}

	var u dbUser
	u.fromUser(newUser)
	if err := dbh(ctx).Insert(&u); err != nil {
		if strings.Contains(err.Error(), `duplicate key value violates unique constraint "users_login"`) {
			return nil, &store.AccountAlreadyExistsError{Login: newUser.Login, UID: newUser.UID}
		}
		return nil, err
	}
	return u.toUser(), nil
}

func (s *Accounts) Update(ctx context.Context, modUser *sourcegraph.User) error {
	var u dbUser
	u.fromUser(modUser)
	if _, err := dbh(ctx).Update(&u); err != nil {
		return err
	}
	return nil
}

func init() {
	Schema.Map.AddTableWithName(PasswordResetRequest{}, "password_reset_requests").SetKeys(false, "Token")
}

type PasswordResetRequest struct {
	Token string
	UID   int32
}

func (s *Accounts) RequestPasswordReset(ctx context.Context, user *sourcegraph.User) (*sourcegraph.PasswordResetToken, error) {
	// 62 characters in upper, lower, and decimal, 62^44 is slightly more than
	// 2^256, so it's astronomically hard to guess, but doesn't take an excessive
	// amount of space to store.
	const tokenLength = 44
	if user.UID == 0 {
		return nil, errors.New("UID must be set")
	}
	token := randstring.NewLen(tokenLength)
	req := PasswordResetRequest{
		Token: token,
		UID:   user.UID,
	}
	if err := dbh(ctx).Insert(&req); err != nil {
		return nil, fmt.Errorf("Error saving password reset token: %s", err)
	}
	return &sourcegraph.PasswordResetToken{Token: token}, nil
}

func (s *Accounts) ResetPassword(ctx context.Context, newPass *sourcegraph.NewPassword) error {
	genericErr := errors.New("error reseting password") // don't need to reveal everything
	req := make([]PasswordResetRequest, 0)
	err := dbh(ctx).Select(&req, `SELECT * FROM password_reset_requests WHERE Token=$1`, newPass.Token.Token)
	if err != nil || len(req) != 1 {
		log15.Warn("db", "token does not exist in password reset database:", err)
		return genericErr
	}
	log15.Info("db", "reseting password for", req[0].UID)
	if err := (Password{}).SetPassword(ctx, req[0].UID, newPass.Password); err != nil {
		return fmt.Errorf("Error changing password: %s", err)
	}
	_, err = dbh(ctx).Exec(`DELETE FROM password_reset_requests WHERE Token=$1`, newPass.Token.Token)
	if err != nil {
		log15.Warn("db", "error deleting token", err)
		return nil
	}
	return nil
}
