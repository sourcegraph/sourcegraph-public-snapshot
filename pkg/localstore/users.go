package localstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// matchUsername represents the limitations on Sourcegraph usernames. It is
// based on the limitations GitHub places on their usernames. This pattern is
// canonical, so any frontend or DB username validation should be based on a
// pattern equivalent to this one.
var matchUsername = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,36}[a-zA-Z0-9])?$`)

type ErrUserNotFound struct {
	args []interface{}
}

func (err ErrUserNotFound) Error() string {
	return fmt.Sprintf("user not found: %v", err.args)
}

// users provides access to the `users` table.
//
// For a detailed overview of the schema, see schema.txt.
type users struct{}

func (*users) Create(auth0ID, email, username, displayName string, avatarURL *string) (*sourcegraph.User, error) {
	createdAt := time.Now()
	updatedAt := createdAt
	var id int32
	var avatarURLValue sql.NullString
	if avatarURL != nil {
		avatarURLValue = sql.NullString{String: *avatarURL, Valid: true}
	}
	err := globalDB.QueryRow(
		"INSERT INTO users(auth0_id, email, username, display_name, avatar_url, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		auth0ID, email, username, displayName, avatarURLValue, createdAt, updatedAt).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &sourcegraph.User{
		ID:          id,
		Auth0ID:     auth0ID,
		Email:       email,
		Username:    username,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

func (u *users) Update(id int32, displayName *string, avatarURL *string) (*sourcegraph.User, error) {
	if displayName == nil && avatarURL == nil {
		return nil, errors.New("no update values provided")
	}

	user, err := u.GetByID(id)
	if err != nil {
		return nil, err
	}

	if displayName != nil {
		user.DisplayName = *displayName
		if _, err := globalDB.Exec("UPDATE users SET display_name=$1 WHERE id=$2", user.DisplayName, id); err != nil {
			return nil, err
		}
	}
	if avatarURL != nil {
		user.AvatarURL = avatarURL
		if _, err := globalDB.Exec("UPDATE users SET avatar_url=$1 WHERE id=$2", *user.AvatarURL, id); err != nil {
			return nil, err
		}
	}
	user.UpdatedAt = time.Now()
	if _, err := globalDB.Exec("UPDATE users SET updated_at=$1 WHERE id=$2", user.UpdatedAt, id); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *users) GetByID(id int32) (*sourcegraph.User, error) {
	return u.getOneBySQL("WHERE id=$1 AND deleted_at IS NULL LIMIT 1", id)
}

func (u *users) GetByCurrentAuthUser(ctx context.Context) (*sourcegraph.User, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("no current user")
	}

	users, err := u.getBySQL("WHERE auth0_id=$1 AND deleted_at IS NULL LIMIT 1", actor.UID)
	if err != nil || len(users) == 0 {
		return nil, err
	}
	return users[0], nil
}

func (u *users) getOneBySQL(query string, args ...interface{}) (*sourcegraph.User, error) {
	users, err := u.getBySQL(query, args...)
	if err != nil {
		return nil, err
	}
	if len(users) != 1 {
		return nil, ErrUserNotFound{args}
	}
	return users[0], nil
}

// getBySQL returns users matching the SQL query, if any exist.
func (*users) getBySQL(query string, args ...interface{}) ([]*sourcegraph.User, error) {
	rows, err := globalDB.Query("SELECT id, auth0_id, email, username, display_name, avatar_url, created_at, updated_at FROM users "+query, args...)
	if err != nil {
		return nil, err
	}

	users := []*sourcegraph.User{}
	defer rows.Close()
	for rows.Next() {
		var u sourcegraph.User
		var avatarUrl sql.NullString
		err := rows.Scan(&u.ID, &u.Auth0ID, &u.Email, &u.Username, &u.DisplayName, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if avatarUrl.Valid {
			u.AvatarURL = &avatarUrl.String
		}
		users = append(users, &u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
