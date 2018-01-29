package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// UserEmail represents a row in the `user_emails` table.
type UserEmail struct {
	UserID           int32
	Email            string
	CreatedAt        time.Time
	VerificationCode *string
	VerifiedAt       *time.Time
}

// userEmails provides access to the `user_emails` table.
type userEmails struct{}

func (*userEmails) GetEmail(ctx context.Context, id int32) (email string, verified bool, err error) {
	if Mocks.UserEmails.GetEmail != nil {
		return Mocks.UserEmails.GetEmail(ctx, id)
	}

	if err := globalDB.QueryRowContext(ctx, "SELECT email, verified_at IS NOT NULL AS verified FROM user_emails WHERE user_id=$1 ORDER BY created_at ASC, email ASC LIMIT 1",
		id,
	).Scan(&email, &verified); err != nil {
		return "", false, userNotFoundErr{[]interface{}{fmt.Sprintf("id %d", id)}}
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

// getBySQL returns user emails matching the SQL query, if any exist.
func (*userEmails) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*UserEmail, error) {
	rows, err := globalDB.QueryContext(ctx,
		`SELECT user_emails.user_id, user_emails.email, user_emails.created_at, user_emails.verification_code,
				user_emails.verified_at FROM user_emails `+query, args...)
	if err != nil {
		return nil, err
	}

	var userEmails []*UserEmail
	defer rows.Close()
	for rows.Next() {
		var v UserEmail
		err := rows.Scan(&v.UserID, &v.Email, &v.CreatedAt, &v.VerificationCode, &v.VerifiedAt)
		if err != nil {
			return nil, err
		}
		userEmails = append(userEmails, &v)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return userEmails, nil
}

func (*userEmails) ListByUser(ctx context.Context, userID int32) ([]*UserEmail, error) {
	return (&userEmails{}).getBySQL(ctx, "WHERE user_id=$1 ORDER BY created_at ASC, email ASC", userID)
}
