package localstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	"golang.org/x/crypto/bcrypt"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/lib/pq"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// UserNamePattern represents the limitations on Sourcegraph usernames. It is
// based on the limitations GitHub places on their usernames. This pattern is
// canonical, so any frontend or DB username validation should be based on a
// pattern equivalent to this one.
const UsernamePattern = `[a-zA-Z0-9]([a-zA-Z0-9-]{0,36}[a-zA-Z0-9])?`

var MatchUsernameString = regexp.MustCompile("^" + UsernamePattern + "$")

// users provides access to the `users` table.
//
// For a detailed overview of the schema, see schema.txt.
type users struct{}

// ErrUserNotFound is the error that is returned when
// a user is not found.
type ErrUserNotFound struct {
	args []interface{}
}

func (err ErrUserNotFound) Error() string {
	return fmt.Sprintf("user not found: %v", err.args)
}

// ErrCannotCreateUser is the error that is returned when
// a user cannot be added to the DB due to a constraint.
type ErrCannotCreateUser struct {
	code string
}

func (err ErrCannotCreateUser) Error() string {
	return fmt.Sprintf("cannot create user: %v", err.code)
}

func (err ErrCannotCreateUser) Code() string {
	return err.code
}

func (*users) Create(ctx context.Context, auth0ID, email, username, displayName, provider string, avatarURL *string, password string, emailCode string) (newUser *sourcegraph.User, err error) {
	createdAt := time.Now()
	updatedAt := createdAt
	var id int32
	var avatarURLValue sql.NullString
	if avatarURL != nil {
		avatarURLValue = sql.NullString{String: *avatarURL, Valid: true}
	}

	var passwd sql.NullString
	if password == "" {
		passwd = sql.NullString{Valid: false}
	} else {
		// Compute hash of password
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		passwd = sql.NullString{Valid: true, String: string(hash)}
	}

	dbEmailCode := sql.NullString{String: emailCode}
	dbEmailCode.Valid = emailCode == ""

	// Wrap in transaction so we can execute hooks below atomically.
	tx, err := globalDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				err = multierror.Append(err, rollErr)
			}
			return
		}
		err = tx.Commit()
	}()

	err = tx.QueryRowContext(
		ctx,
		"INSERT INTO users(auth_id, email, username, display_name, provider, avatar_url, created_at, updated_at, passwd, email_code) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id",
		auth0ID, email, username, displayName, provider, avatarURLValue, createdAt, updatedAt, passwd, emailCode).Scan(&id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Constraint {
			case "users_username_key":
				return nil, ErrCannotCreateUser{"err_username_exists"}
			case "users_email_key":
				return nil, ErrCannotCreateUser{"err_email_exists"}
			case "users_auth_id_key":
				return nil, ErrCannotCreateUser{"err_auth_id_exists"}
			}
		}
		return nil, err
	}

	{
		// Run hooks.
		//
		// NOTE: If we need more hooks in the future, we should do something better than just
		// adding random calls here.

		// Ensure the user (all users, actually) is joined to the orgs specified in auth.userOrgMap.
		orgs, errs := conf.Get().Auth.UserOrgMap.OrgsForAllUsersToJoin()
		for _, err := range errs {
			log15.Warn(err.Error())
		}
		if err := OrgMembers.CreateMembershipInOrgsForAllUsers(ctx, tx, orgs); err != nil {
			return nil, err
		}
	}

	return &sourcegraph.User{
		ID:          id,
		Auth0ID:     auth0ID,
		Email:       email,
		Username:    username,
		DisplayName: displayName,
		Provider:    provider,
		AvatarURL:   avatarURL,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

func (u *users) Update(ctx context.Context, id int32, username *string, displayName *string, avatarURL *string) (*sourcegraph.User, error) {
	if username == nil && displayName == nil && avatarURL == nil {
		return nil, errors.New("no update values provided")
	}

	user, err := u.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if username != nil {
		user.Username = *username
		if _, err := globalDB.ExecContext(ctx, "UPDATE users SET username=$1 WHERE id=$2", user.Username, id); err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Constraint == "users_username_key" {
					return nil, errors.New("username already exists")
				}
				return nil, err
			}
		}
	}
	if displayName != nil {
		user.DisplayName = *displayName
		if _, err := globalDB.ExecContext(ctx, "UPDATE users SET display_name=$1 WHERE id=$2", user.DisplayName, id); err != nil {
			return nil, err
		}
	}
	if avatarURL != nil {
		user.AvatarURL = avatarURL
		if _, err := globalDB.ExecContext(ctx, "UPDATE users SET avatar_url=$1 WHERE id=$2", *user.AvatarURL, id); err != nil {
			return nil, err
		}
	}
	user.UpdatedAt = time.Now()
	if _, err := globalDB.ExecContext(ctx, "UPDATE users SET updated_at=$1 WHERE id=$2", user.UpdatedAt, id); err != nil {
		return nil, err
	}

	return user, nil
}

// CheckAndDecrementInviteQuota should be called before the user (identified by userID) is
// allowed to invite any other user. If err != nil, then the user is not allowed to invite
// any other user (either because they've invited too many users, or some other error
// occurred). If the user has quota remaining, their quota is decremented.
func (u *users) CheckAndDecrementInviteQuota(ctx context.Context, userID int32) error {
	var quotaRemaining int32
	sqlQuery := `
	UPDATE users SET invite_quota=(invite_quota - 1)
	WHERE users.id=$1 AND invite_quota>0 AND deleted_at IS NULL
	RETURNING invite_quota`
	row := globalDB.QueryRowContext(ctx, sqlQuery, userID)
	if err := row.Scan(&quotaRemaining); err == sql.ErrNoRows {
		// It's possible that some other problem occurred, such as the user being deleted,
		// but treat that as a quota exceeded error, too.
		return ErrInviteQuotaExceeded
	} else if err != nil {
		return err
	}
	return nil // the user has remaining quota to send invites
}

// ErrInviteQuotaExceeded indicates that the user has exceeded their invite quota
// and may not send any more invites.
var ErrInviteQuotaExceeded = errors.New("invite quota exceeded")

func (u *users) GetByID(ctx context.Context, id int32) (*sourcegraph.User, error) {
	if Mocks.Users.GetByID != nil {
		return Mocks.Users.GetByID(ctx, id)
	}
	return u.getOneBySQL(ctx, "WHERE id=$1 AND deleted_at IS NULL LIMIT 1", id)
}

func (u *users) GetByAuth0ID(ctx context.Context, id string) (*sourcegraph.User, error) {
	if Mocks.Users.GetByAuth0ID != nil {
		return Mocks.Users.GetByAuth0ID(ctx, id)
	}
	return u.getOneBySQL(ctx, "WHERE auth_id=$1 AND deleted_at IS NULL LIMIT 1", id)
}

func (u *users) GetByEmail(ctx context.Context, email string) (*sourcegraph.User, error) {
	return u.getOneBySQL(ctx, "WHERE email=$1 AND deleted_at IS NULL LIMIT 1", email)
}

func (u *users) GetByUsername(ctx context.Context, username string) (*sourcegraph.User, error) {
	return u.getOneBySQL(ctx, "WHERE username=$1 AND deleted_at IS NULL LIMIT 1", username)
}

func (u *users) GetByCurrentAuthUser(ctx context.Context) (*sourcegraph.User, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("no current user")
	}

	return u.getOneBySQL(ctx, "WHERE auth_id=$1 AND deleted_at IS NULL LIMIT 1", actor.UID)
}

// ListByOrg returns users for a given org. It can also query a list of specific
// users by either auth0IDs or usernames.
func (u *users) ListByOrg(ctx context.Context, orgID int32, auth0IDs, usernames []string) ([]*sourcegraph.User, error) {
	if Mocks.Users.ListByOrg != nil {
		return Mocks.Users.ListByOrg(ctx, orgID, auth0IDs, usernames)
	}
	conds := []*sqlf.Query{}
	filters := []*sqlf.Query{}
	if len(auth0IDs) > 0 {
		items := []*sqlf.Query{}
		for _, id := range auth0IDs {
			items = append(items, sqlf.Sprintf("%s", id))
		}
		filters = append(filters, sqlf.Sprintf("u.auth_id IN (%s)", sqlf.Join(items, ",")))
	}
	if len(usernames) > 0 {
		items := []*sqlf.Query{}
		for _, u := range usernames {
			items = append(items, sqlf.Sprintf("%s", u))
		}
		filters = append(filters, sqlf.Sprintf("u.username IN (%s)", sqlf.Join(items, ",")))
	}
	if len(filters) > 0 {
		conds = append(conds, sqlf.Sprintf("(%s)", sqlf.Join(filters, "OR")))
	}
	conds = append(conds, sqlf.Sprintf("org_members.org_id=%d", orgID), sqlf.Sprintf("u.deleted_at IS NULL"))
	q := sqlf.Sprintf("JOIN org_members ON (org_members.user_id = u.auth_id) WHERE %s", sqlf.Join(conds, "AND"))
	return u.getBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
}

func (u *users) List(ctx context.Context) ([]*sourcegraph.User, error) {
	return u.getBySQL(ctx, "")
}

func (u *users) getOneBySQL(ctx context.Context, query string, args ...interface{}) (*sourcegraph.User, error) {
	users, err := u.getBySQL(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(users) != 1 {
		return nil, ErrUserNotFound{args}
	}
	return users[0], nil
}

// getBySQL returns users matching the SQL query, if any exist.
func (*users) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.User, error) {
	rows, err := globalDB.QueryContext(ctx, "SELECT u.id, u.auth_id, u.email, u.username, u.display_name, u.provider, u.avatar_url, u.created_at, u.updated_at, u.email_code is null FROM users u "+query, args...)
	if err != nil {
		return nil, err
	}

	users := []*sourcegraph.User{}
	defer rows.Close()
	for rows.Next() {
		var u sourcegraph.User
		var avatarUrl sql.NullString
		err := rows.Scan(&u.ID, &u.Auth0ID, &u.Email, &u.Username, &u.DisplayName, &u.Provider, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt, &u.Verified)
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

func (u *users) IsPassword(ctx context.Context, id int32, password string) (bool, error) {
	var passwd sql.NullString
	if err := globalDB.QueryRowContext(ctx, "SELECT passwd FROM users WHERE id=$1", id).Scan(&passwd); err != nil {
		return false, err
	}
	if !passwd.Valid {
		return false, nil
	}
	return bcrypt.CompareHashAndPassword([]byte(passwd.String), []byte(password)) == nil, nil
}

func (u *users) ValidateEmail(ctx context.Context, id int32, userCode string) (bool, error) {
	var dbCode sql.NullString
	if err := globalDB.QueryRowContext(ctx, "SELECT email_code FROM users WHERE id=$1", id).Scan(&dbCode); err != nil {
		return false, err
	}
	if !dbCode.Valid {
		return false, errors.New("email already verified")
	}
	if dbCode.String != userCode {
		return false, nil
	}
	if _, err := globalDB.ExecContext(ctx, "UPDATE users SET email_code=null WHERE id=$1", id); err != nil {
		return false, err
	}
	return true, nil
}
