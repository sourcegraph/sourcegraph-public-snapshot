package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// userEmails provides access to the `user_emails` table.
type userEmails struct{}

func (*userEmails) GetEmail(ctx context.Context, id int32) (email string, verified bool, err error) {
	if Mocks.UserEmails.GetEmail != nil {
		return Mocks.UserEmails.GetEmail(ctx, id)
	}

	if err := globalDB.QueryRowContext(ctx, "SELECT email, verified_at IS NOT NULL AS verified FROM user_emails WHERE user_id=$1 ORDER BY created_at ASC, email ASC LIMIT 1",
		id,
	).Scan(&email, &verified); err != nil {
		return "", false, ErrUserNotFound{[]interface{}{fmt.Sprintf("id %d", id)}}
	}
	return email, verified, nil
}

func (*userEmails) ValidateEmail(ctx context.Context, id int32, userCode string) (bool, error) {
	var dbCode sql.NullString
	if err := globalDB.QueryRowContext(ctx, "SELECT verification_code FROM user_emails WHERE user_id=$1", id).Scan(&dbCode); err != nil {
		return false, err
	}
	if !dbCode.Valid {
		return false, errors.New("email already verified")
	}
	if dbCode.String != userCode {
		return false, nil
	}
	if _, err := globalDB.ExecContext(ctx, "UPDATE user_emails SET verification_code=null, verified_at=now() WHERE user_id=$1", id); err != nil {
		return false, err
	}
	return true, nil
}
