package localstore

import (
	"database/sql"

	"golang.org/x/net/context"
	"gopkg.in/gorp.v1"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/util/dbutil"
)

type userEmailAddrRow struct {
	UID int
	sourcegraph.EmailAddr
}

func init() {
	AppSchema.Map.AddTableWithName(userEmailAddrRow{}, "user_email").SetKeys(false, "UID", "Email")
	AppSchema.CreateSQL = append(AppSchema.CreateSQL,
		`ALTER TABLE user_email ALTER COLUMN email TYPE citext;`,
		`CREATE INDEX user_email_email ON user_email(email) WHERE (NOT blacklisted);`,
		`CREATE UNIQUE INDEX user_email_email_primary ON user_email(email, "primary") WHERE (NOT blacklisted);`,
	)
}

func (s *users) GetWithEmail(ctx context.Context, emailAddr sourcegraph.EmailAddr) (*sourcegraph.User, error) {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Users.GetWithEmail"); err != nil {
		return nil, err
	}
	var emailAddrRow userEmailAddrRow
	query := `SELECT * FROM user_email WHERE email=$1 AND "primary"=true`
	if err := appDBH(ctx).SelectOne(&emailAddrRow, query, emailAddr.Email); err == sql.ErrNoRows {
		return nil, &store.UserNotFoundError{Email: emailAddr.Email}
	} else if err != nil {
		return nil, err
	}
	return s.getByUID(ctx, emailAddrRow.UID)
}

func (s *users) ListEmails(ctx context.Context, user sourcegraph.UserSpec) ([]*sourcegraph.EmailAddr, error) {
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "Users.ListEmails", user.UID); err != nil {
		return nil, err
	}
	if user.UID == 0 {
		return nil, &store.UserNotFoundError{UID: 0}
	}

	var emailAddrRows []*userEmailAddrRow
	sql := `SELECT * FROM user_email WHERE uid=$1 ORDER BY "primary" DESC, verified DESC`
	if _, err := appDBH(ctx).Select(&emailAddrRows, sql, user.UID); err != nil {
		return nil, err
	}

	emailAddrs := make([]*sourcegraph.EmailAddr, len(emailAddrRows))
	for i, row := range emailAddrRows {
		emailAddrs[i] = &row.EmailAddr
	}
	return emailAddrs, nil
}

func (s *accounts) UpdateEmails(ctx context.Context, user sourcegraph.UserSpec, emails []*sourcegraph.EmailAddr) error {
	if err := accesscontrol.VerifyUserSelfOrAdmin(ctx, "Accounts.UpdateEmails", user.UID); err != nil {
		return err
	}
	if user.UID == 0 {
		return &store.UserNotFoundError{UID: 0}
	}

	return dbutil.Transact(appDBH(ctx), func(tx gorp.SqlExecutor) error {
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
