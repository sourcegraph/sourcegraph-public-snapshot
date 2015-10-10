package pgsql

import (
	"github.com/sqs/modl"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/dbutil"
)

type userEmailAddrRow struct {
	UID int
	sourcegraph.EmailAddr
}

func init() {
	Schema.Map.AddTableWithName(userEmailAddrRow{}, "user_email").SetKeys(false, "UID", "Email")
	Schema.CreateSQL = append(Schema.CreateSQL,
		`ALTER TABLE user_email ALTER COLUMN email TYPE citext;`,
		`CREATE INDEX user_email_email ON user_email(email) WHERE (NOT blacklisted);`,
		`CREATE UNIQUE INDEX user_email_email_primary ON user_email(email, "primary") WHERE (NOT blacklisted);`,
	)
}

func (s *Users) GetWithEmail(ctx context.Context, emailAddr sourcegraph.EmailAddr) (*sourcegraph.User, error) {
	var emailAddrRows []*userEmailAddrRow
	sql := `SELECT * FROM user_email WHERE email=$1 AND "primary"=true`
	if err := dbh(ctx).Select(&emailAddrRows, sql, emailAddr.Email); err != nil {
		return nil, err
	}
	if len(emailAddrRows) == 0 {
		return nil, &store.UserNotFoundError{Email: emailAddr.Email}
	}
	return s.getByUID(ctx, emailAddrRows[0].UID)
}

func (s *Users) ListEmails(ctx context.Context, user sourcegraph.UserSpec) ([]*sourcegraph.EmailAddr, error) {
	if user.UID == 0 {
		panic("UID == 0")
	}

	var emailAddrRows []*userEmailAddrRow
	sql := `SELECT * FROM user_email WHERE uid=$1 ORDER BY "primary" DESC, verified DESC`
	if err := dbh(ctx).Select(&emailAddrRows, sql, user.UID); err != nil {
		return nil, err
	}

	emailAddrs := make([]*sourcegraph.EmailAddr, len(emailAddrRows))
	for i, row := range emailAddrRows {
		emailAddrs[i] = &row.EmailAddr
	}
	return emailAddrs, nil
}

func (s *Accounts) UpdateEmails(ctx context.Context, user sourcegraph.UserSpec, emails []*sourcegraph.EmailAddr) error {
	if user.UID == 0 {
		panic("UID == 0")
	}

	return dbutil.Transact(dbh(ctx), func(tx modl.SqlExecutor) error {
		// Clear out all existing from DB, and add in the merged (final) list.
		if _, err := tx.Exec(`DELETE FROM user_email WHERE uid=$1;`, user.UID); err != nil {
			return err
		}

		for _, email := range emails {
			if err := tx.Insert(&userEmailAddrRow{UID: int(user.UID), EmailAddr: *email}); err != nil {
				return err
			}
		}
		return nil
	})
}
