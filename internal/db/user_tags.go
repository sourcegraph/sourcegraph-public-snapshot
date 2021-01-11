package db

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
)

const (
	// If the owner of an external service has this tag, the service is allowed to sync private code
	TagAllowUserExternalServicePrivate = "AllowUserExternalServicePrivate"
	// If the owner of an external service has this tag, the service is allowed to sync public code only
	TagAllowUserExternalServicePublic = "AllowUserExternalServicePublic"
)

// SetTag adds (present=true) or removes (present=false) a tag from the given user's set of tags. An
// error occurs if the user does not exist. Adding a duplicate tag or removing a nonexistent tag is
// not an error.
func (u *UserStore) SetTag(ctx context.Context, userID int32, tag string, present bool) error {
	u.ensureStore()

	var q *sqlf.Query
	if present {
		// Add tag.
		q = sqlf.Sprintf(`UPDATE users SET tags=CASE WHEN NOT %s::text = ANY(tags) THEN (tags || %s::text) ELSE tags END WHERE id=%s`, tag, tag, userID)
	} else {
		// Remove tag.
		q = sqlf.Sprintf(`UPDATE users SET tags=array_remove(tags, %s::text) WHERE id=%s`, tag, userID)
	}

	res, err := u.ExecResult(ctx, q)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return userNotFoundErr{args: []interface{}{userID}}
	}
	return nil
}

// HasTag reports whether the context actor has the given tag.
// If not, it returns false and a nil error.
func (u *UserStore) HasTag(ctx context.Context, userID int32, tag string) (bool, error) {
	if Mocks.Users.HasTag != nil {
		return Mocks.Users.HasTag(ctx, userID, tag)
	}
	u.ensureStore()

	var tags []string
	err := u.QueryRow(ctx, sqlf.Sprintf("SELECT tags FROM users WHERE id = %s", userID)).Scan(pq.Array(&tags))
	if err != nil {
		if err == sql.ErrNoRows {
			return false, userNotFoundErr{[]interface{}{userID}}
		}
		return false, err
	}

	for _, t := range tags {
		if t == tag {
			return true, nil
		}
	}
	return false, nil
}
