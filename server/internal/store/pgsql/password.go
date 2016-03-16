package pgsql

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/server/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
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
	Schema.Map.AddTableWithName(dbPassword{}, tableName).SetKeys(false, "UID")
}

// CheckUIDPassword returns an error if the password argument is not correct for
// the user.
func (p password) CheckUIDPassword(ctx context.Context, UID int32, password string) error {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Password.CheckUIDPassword"); err != nil {
		return err
	}
	hashed, err := dbh(ctx).SelectStr("SELECT hashedpassword FROM passwords WHERE uid=$1;", UID)
	if err != nil {
		return err
	}
	if hashed == "" {
		return &store.UserNotFoundError{UID: int(UID)}
	}
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

func (p password) SetPassword(ctx context.Context, uid int32, password string) error {
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "Password.SetPassword", uid); err != nil {
		return err
	}
	if password == "" {
		return errors.New("password must not be empty")
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
	_, err = dbh(ctx).Exec(query, uid, hashed)
	return err
}
