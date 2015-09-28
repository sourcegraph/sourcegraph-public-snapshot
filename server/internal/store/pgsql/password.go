package pgsql

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/store"
)

// Password is a pgsql backed implementation of the passwords store.
type Password struct{}

var _ store.Password = (*Password)(nil)

type dbPassword struct {
	UID            int32
	HashedPassword []byte
}

var tableName = "passwords"

func init() {
	t := Schema.Map.AddTableWithName(dbPassword{}, tableName)
	t.SetKeys(false, "uid")
}

// CheckUIDPassword returns an error if the password argument is not correct for
// the user.
func (p Password) CheckUIDPassword(ctx context.Context, UID int32, password string) error {
	var records [][]byte
	err := dbh(ctx).Select(&records, "SELECT hashedpassword FROM passwords WHERE uid=$1;", UID)
	if err != nil {
		return err
	}
	if len(records) != 1 {
		return &store.UserNotFoundError{UID: int(UID)}
	}
	hashed := records[0]
	return bcrypt.CompareHashAndPassword(hashed, []byte(password))
}

func (p Password) SetPassword(ctx context.Context, uid int32, password string) error {
	if password == "" {
		return errors.New("password must not be empty")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	if err != nil {
		return err
	}

	var records []int32
	err = dbh(ctx).Select(&records, "SELECT uid FROM passwords WHERE uid=$1;", uid)
	if err != nil {
		return err
	}

	u := dbPassword{UID: uid, HashedPassword: hashed}
	if len(records) == 0 {
		err = dbh(ctx).Insert(&u)
	} else {
		_, err = dbh(ctx).Update(&u)
	}
	if err != nil {
		return err
	}
	return nil
}
