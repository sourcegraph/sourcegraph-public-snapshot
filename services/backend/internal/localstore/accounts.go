package localstore

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"gopkg.in/gorp.v1"
	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/amortize"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
)

// accounts is a DB-backed implementation of the Accounts store.
type accounts struct{}

var _ store.Accounts = (*accounts)(nil)

func (s *accounts) Create(ctx context.Context, newUser *sourcegraph.User, email *sourcegraph.EmailAddr) (*sourcegraph.User, error) {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Accounts.Create"); err != nil {
		return nil, err
	}
	if newUser.UID != 0 {
		return nil, errors.New("uid already set")
	}
	if newUser.Login == "" {
		return nil, errors.New("login must be set")
	}

	var u dbUser
	u.fromUser(newUser)

	err := dbutil.Transact(appDBH(ctx), func(tx gorp.SqlExecutor) error {
		if err := tx.Insert(&u); err != nil {
			if strings.Contains(err.Error(), `duplicate key value violates unique constraint "users_login"`) {
				return &store.AccountAlreadyExistsError{Login: newUser.Login, UID: newUser.UID}
			}
			return err
		}

		var insertedUser dbUser
		if err := tx.SelectOne(&insertedUser, "SELECT * FROM users WHERE login=$1 LIMIT 1", newUser.Login); err != nil {
			return err
		}

		if email != nil {
			if err := tx.Insert(&userEmailAddrRow{UID: int(insertedUser.UID), EmailAddr: *email}); err != nil {
				return grpc.Errorf(codes.AlreadyExists, "%s has already been registered with another account", email.Email)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return u.toUser(), nil
}

func (s *accounts) Update(ctx context.Context, modUser *sourcegraph.User) error {
	// A user can only update their own record, but an admin can update all records.
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "Accounts.Update", modUser.UID); err != nil {
		return err
	}

	a := authpkg.ActorFromContext(ctx)
	// Only admin users can modify access levels of a user.
	if !a.HasAdminAccess() && (modUser.Admin || (a.HasWriteAccess() != modUser.Write)) {
		return grpc.Errorf(codes.PermissionDenied, "need admin privileges to modify user permissions")
	}

	var u dbUser
	u.fromUser(modUser)
	if _, err := appDBH(ctx).Update(&u); err != nil {
		return err
	}
	return nil
}

func (s *accounts) Delete(ctx context.Context, uid int32) error {
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "Accounts.Delete", uid); err != nil {
		return err
	}

	if uid == 0 {
		return &store.UserNotFoundError{UID: 0}
	}

	return dbutil.Transact(appDBH(ctx), func(tx gorp.SqlExecutor) error {
		dbUID := int(uid)

		if _, err := tx.Exec(`DELETE FROM users where uid=$1`, dbUID); err != nil {
			return err
		}

		if _, err := tx.Exec(`DELETE FROM user_email WHERE uid=$1;`, dbUID); err != nil {
			return err
		}

		if _, err := tx.Exec(`DELETE FROM ext_auth_token WHERE "user"=$1;`, dbUID); err != nil {
			return err
		}

		return nil
	})
}

func init() {
	AppSchema.Map.AddTableWithName(passwordResetRequest{}, "password_reset_requests").SetKeys(false, "Token")
}

type passwordResetRequest struct {
	Token     string
	UID       int32
	ExpiresAt time.Time `db:"expires_at"`
}

const passwordResetTokenExpiration = 4 * time.Hour

func (s *accounts) RequestPasswordReset(ctx context.Context, user *sourcegraph.User) (*sourcegraph.PasswordResetToken, error) {
	// 1 out of every 1000 times is just an initial guess as to how often we
	// should go through and delete expired password reset requests.
	if amortize.ShouldAmortize(1, 1000) {
		s.cleanExpiredResets(ctx)
	}
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "Accounts.RequestPasswordReset", user.UID); err != nil {
		return nil, err
	}
	// 62 characters in upper, lower, and decimal, 62^44 is slightly more than
	// 2^256, so it's astronomically hard to guess, but doesn't take an excessive
	// amount of space to store.
	const tokenLength = 44
	if user.UID == 0 {
		return nil, errors.New("UID must be set")
	}
	token := randstring.NewLen(tokenLength)
	expiration := time.Now().Add(passwordResetTokenExpiration)
	req := passwordResetRequest{
		Token:     token,
		UID:       user.UID,
		ExpiresAt: expiration,
	}
	if err := appDBH(ctx).Insert(&req); err != nil {
		return nil, fmt.Errorf("Error saving password reset token: %s", err)
	}
	return &sourcegraph.PasswordResetToken{Token: token}, nil
}

func unmarshalResetRequest(ctx context.Context, token string) (passwordResetRequest, error) {
	var req passwordResetRequest
	err := appDBH(ctx).SelectOne(&req, `SELECT * FROM password_reset_requests WHERE token=$1`, token)
	return req, err
}

func (s *accounts) ResetPassword(ctx context.Context, newPass *sourcegraph.NewPassword) error {
	genericErr := grpc.Errorf(codes.InvalidArgument, "error reseting password") // don't need to reveal everything
	req, err := unmarshalResetRequest(ctx, newPass.Token.Token)
	if err == sql.ErrNoRows {
		log15.Warn("Token does not exist in password reset database", "store", "Accounts", "error", err)
		return genericErr
	} else if err != nil {
		return genericErr
	}
	// We round to the microsecond because that is the maximum resolution
	// allowed by our DB.
	if time.Now().Round(time.Microsecond).Before(req.ExpiresAt.Round(time.Microsecond)) {
		log15.Info("Resetting password", "store", "Accounts", "UID", req.UID)
		if err := (password{}).SetPassword(ctx, req.UID, newPass.Password); err != nil {
			return fmt.Errorf("error changing password: %s", err)
		}
		// This changes the expiry date to some point in the past.
		req.ExpiresAt = time.Time{}
		_, err := appDBH(ctx).Update(&req)
		if err != nil {
			log15.Warn("Error updating token", "store", "Accounts", "error", err)
		}
		return nil
	}
	return grpc.Errorf(codes.InvalidArgument, "password reset token has expired")
}

// cleanExpiredResets deletes password reset requests from the DB whose expiration date has passed.
func (s *accounts) cleanExpiredResets(ctx context.Context) error {
	_, err := appDBH(ctx).Exec(`DELETE FROM password_reset_requests WHERE expires_at < now()`)
	if err != nil {
		return fmt.Errorf("error when cleaning up expired password resets: %s", err)
	}
	return nil
}
