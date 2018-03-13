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

// userEmailNotFoundError is the error that is returned when a user email is not found.
type userEmailNotFoundError struct {
	args []interface{}
}

func (err userEmailNotFoundError) Error() string {
	return fmt.Sprintf("user email not found: %v", err.args)
}

func (err userEmailNotFoundError) NotFound() bool {
	return true
}

// userEmails provides access to the `user_emails` table.
type userEmails struct{}

// GetPrimaryEmail gets the oldest email associated with the user.
func (*userEmails) GetPrimaryEmail(ctx context.Context, id int32) (email string, verified bool, err error) {
	if Mocks.UserEmails.GetPrimaryEmail != nil {
		return Mocks.UserEmails.GetPrimaryEmail(ctx, id)
	}

	if err := globalDB.QueryRowContext(ctx, "SELECT email, verified_at IS NOT NULL AS verified FROM user_emails WHERE user_id=$1 ORDER BY created_at ASC, email ASC LIMIT 1",
		id,
	).Scan(&email, &verified); err != nil {
		return "", false, userNotFoundErr{[]interface{}{fmt.Sprintf("id %d", id)}}
	}
	return email, verified, nil
}

// Get gets information about the user's associated email address.
func (*userEmails) Get(ctx context.Context, userID int32, email string) (emailCanonicalCase string, verified bool, err error) {
	if Mocks.UserEmails.Get != nil {
		return Mocks.UserEmails.Get(userID, email)
	}

	if err := globalDB.QueryRowContext(ctx, "SELECT email, verified_at IS NOT NULL AS verified FROM user_emails WHERE user_id=$1 AND email=$2",
		userID, email,
	).Scan(&emailCanonicalCase, &verified); err != nil {
		return "", false, userEmailNotFoundError{[]interface{}{fmt.Sprintf("userID %d email %q", userID, email)}}
	}
	return emailCanonicalCase, verified, nil
}

// Add adds new user email. When added, it is always unverified.
func (*userEmails) Add(ctx context.Context, userID int32, email string, verificationCode *string) error {
	_, err := globalDB.ExecContext(ctx, "INSERT INTO user_emails(user_id, email, verification_code) VALUES($1, $2, $3)", userID, email, verificationCode)
	return err
}

// Remove removes a user email. It returns an error if there is no such email associated with the user.
func (*userEmails) Remove(ctx context.Context, userID int32, email string) error {
	res, err := globalDB.ExecContext(ctx, "DELETE FROM user_emails WHERE user_id=$1 AND email=$2", userID, email)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errors.New("user email not found")
	}
	return nil
}

// Verify verifies the user's primary email address given the email verification code. If the code
// is not correct (not the one originally used when creating the user or adding the user email),
// then it returns false.
func (*userEmails) Verify(ctx context.Context, id int32, userCode string) (bool, error) {
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

// SetVerified bypasses the normal email verification code process and manually sets the verified
// status for an email.
func (*userEmails) SetVerified(ctx context.Context, userID int32, email string, verified bool) error {
	var res sql.Result
	var err error
	if verified {
		// Mark as verified.
		res, err = globalDB.ExecContext(ctx, "UPDATE user_emails SET verification_code=null, verified_at=now() WHERE user_id=$1 AND email=$2", userID, email)
	} else {
		// Mark as unverified.
		res, err = globalDB.ExecContext(ctx, "UPDATE user_emails SET verification_code=null, verified_at=null WHERE user_id=$1 AND email=$2", userID, email)
	}
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errors.New("user email not found")
	}
	return nil
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
