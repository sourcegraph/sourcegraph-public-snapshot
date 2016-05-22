package localstore

import (
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/server/accesscontrol"
)

// password is a pgsql backed implementation of the passwords store.
type password struct{}

var _ store.Password = (*password)(nil)

type dbPassword struct {
	UID            int32
	HashedPassword []byte
}

var tableName = "passwords"

func init() {
	AppSchema.Map.AddTableWithName(dbPassword{}, tableName).SetKeys(false, "UID")
}

// CheckUIDPassword returns an error if the password argument is not correct for
// the user.
func (p password) CheckUIDPassword(ctx context.Context, UID int32, password string) error {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Password.CheckUIDPassword"); err != nil {
		return err
	}
	hashed, err := appDBH(ctx).SelectStr("SELECT hashedpassword FROM passwords WHERE uid=$1;", UID)
	if err != nil {
		return err
	}
	if hashed == "" {
		// Either the user has no password (and can only log in via
		// GitHub OAuth2, etc.) or the user does not exist.
		return grpc.Errorf(codes.PermissionDenied, "password login not allowed for uid %d", UID)
	}
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

func (p password) SetPassword(ctx context.Context, uid int32, password string) error {
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "Password.SetPassword", uid); err != nil {
		return err
	}

	if password == "" {
		// Clear password (user can only log in via GitHub OAuth2, for
		// example).
		_, err := appDBH(ctx).Exec(`DELETE FROM passwords WHERE uid=$1;`, uid)
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	if err != nil {
		return err
	}

	query := `
WITH upsert AS (
  UPDATE passwords SET hashedpassword=$2 WHERE uid=$1 RETURNING *
)
INSERT INTO passwords(uid, hashedpassword) SELECT $1, $2 WHERE NOT EXISTS (SELECT * FROM upsert);`
	_, err = appDBH(ctx).Exec(query, uid, hashed)
	return err
}
