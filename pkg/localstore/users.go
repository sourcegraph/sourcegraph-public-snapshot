package localstore

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	"golang.org/x/crypto/bcrypt"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/lib/pq"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore/auth0tmp"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// UserNamePattern represents the limitations on Sourcegraph usernames. It is
// based on the limitations GitHub places on their usernames. This pattern is
// canonical, so any frontend or DB username validation should be based on a
// pattern equivalent to this one.
const UsernamePattern = `[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])){0,38}`

var MatchUsernameString = regexp2.MustCompile("^"+UsernamePattern+"$", 0)

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

// Create creates a new user in the database. The provider specifies what identity providers was responsible for authenticating
// the user:
// - If the provider is "", the user was authenticated the native-auth UI using Auth0
// - If the provider is "native", the user was authenticated by the native-auth UI using native authentication
// - If the provider is something else, the user was authenticated by an SSO provider
//
// Native-auth users must also specify a password and email verification code upon creation. When the user's email is
// verified, the email verification code is set to null in the DB. All other users (including Auth0
// Auth0 users that were authenticated using the native-auth UI) have a null password and email verification code.
func (*users) Create(ctx context.Context, auth0ID, email, username, displayName, provider string, avatarURL *string, password string, emailCode string) (newUser *sourcegraph.User, err error) {
	if provider == sourcegraph.UserProviderNative && (password == "" || emailCode == "") {
		return nil, errors.New("no password or email code provided for new native-auth user")
	}
	if provider != sourcegraph.UserProviderNative && (password != "" || emailCode != "") {
		return nil, errors.New("password and/or email verification code provided for non-native users")
	}

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
		passwd, err = hashPassword(password)
		if err != nil {
			return nil, err
		}
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

	// Make this user a site admin if they're the first user.
	const makeSiteAdminSQLExpr = `(SELECT COUNT(*) FROM users) = 0`
	var isSiteAdmin bool

	err = tx.QueryRowContext(
		ctx,
		"INSERT INTO users(auth_id, email, username, display_name, provider, avatar_url, created_at, updated_at, passwd, email_code, site_admin) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, "+makeSiteAdminSQLExpr+") RETURNING id, site_admin",
		auth0ID, email, username, displayName, provider, avatarURLValue, createdAt, updatedAt, passwd, emailCode).Scan(&id, &isSiteAdmin)
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
		orgs, errs := orgsForAllUsersToJoin(conf.Get().AuthUserOrgMap)
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
		SiteAdmin:   isSiteAdmin,
	}, nil
}

// orgsForAllUsersToJoin returns the list of org names that all users should be joined to. The second return value
// is a list of errors encountered while generating this list. Note that even if errors are returned, the first
// return value is still valid.
func orgsForAllUsersToJoin(userOrgMap map[string][]string) ([]string, []error) {
	var errors []error
	for userPattern, orgs := range userOrgMap {
		if userPattern != "*" {
			errors = append(errors, fmt.Errorf("unsupported auth.userOrgMap user pattern %q (only \"*\" is supported)", userPattern))
			continue
		}
		return orgs, errors
	}
	return nil, errors
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

func (u *users) SetIsSiteAdmin(ctx context.Context, id int32, isSiteAdmin bool) error {
	_, err := globalDB.ExecContext(ctx, "UPDATE users SET site_admin=$1 WHERE id=$2", isSiteAdmin, id)
	return err
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
	if Mocks.Users.GetByCurrentAuthUser != nil {
		return Mocks.Users.GetByCurrentAuthUser(ctx)
	}

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
	return u.getBySQL(ctx, "ORDER BY id ASC")
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
	rows, err := globalDB.QueryContext(ctx, "SELECT u.id, u.auth_id, u.email, u.username, u.display_name, u.provider, u.avatar_url, u.created_at, u.updated_at, u.email_code is null, u.site_admin FROM users u "+query, args...)
	if err != nil {
		return nil, err
	}

	users := []*sourcegraph.User{}
	defer rows.Close()
	for rows.Next() {
		var u sourcegraph.User
		var avatarUrl sql.NullString
		err := rows.Scan(&u.ID, &u.Auth0ID, &u.Email, &u.Username, &u.DisplayName, &u.Provider, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt, &u.Verified, &u.SiteAdmin)
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
	if conf.AuthProvider() == "auth0" {
		user, err := u.GetByID(ctx, id)
		if err != nil {
			return false, err
		}
		// During the transition, new users will have provider=="native" and no "auth0|" prefix.
		// We need to check those in our own DB.
		if user.Provider == "auth0" || strings.HasPrefix(user.Auth0ID, "auth0|") {
			ok, err := auth0tmp.CheckPassword(ctx, user.Email, password)
			// log15.Info("checking password via auth0", "user", user.Username, "auth0ID", user.Auth0ID, "email", user.Email, "ok", ok, "err", err)
			return ok, err
		}
	}

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

var (
	passwordResetRateLimit    = "1 minute"
	ErrPasswordResetRateLimit = errors.New("password reset rate limit reached")
)

func (u *users) RenewPasswordResetCode(ctx context.Context, id int32) (string, error) {
	if _, err := u.GetByID(ctx, id); err != nil {
		return "", err
	}
	var b [40]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	code := base64.StdEncoding.EncodeToString(b[:])
	res, err := globalDB.ExecContext(ctx, "UPDATE users SET passwd_reset_code=$1, passwd_reset_time=now() WHERE id=$2 AND (passwd_reset_time IS NULL OR passwd_reset_time + interval '"+passwordResetRateLimit+"' < now())", code, id)
	if err != nil {
		return "", err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return "", err
	}
	if affected == 0 {
		return "", ErrPasswordResetRateLimit
	}

	return code, nil
}

func (u *users) SetPassword(ctx context.Context, id int32, resetCode string, newPassword string) (bool, error) {
	// 🚨 SECURITY: no empty passwords
	if newPassword == "" {
		return false, errors.New("new password was empty")
	}
	// 🚨 SECURITY: check resetCode against what's in the DB and that it's not expired
	r := globalDB.QueryRowContext(ctx, "SELECT count(*) FROM users WHERE id=$1 AND passwd_reset_code=$2 AND passwd_reset_time + interval '4 hours' > now()", id, resetCode)
	var ct int
	if err := r.Scan(&ct); err != nil {
		return false, err
	}
	if ct > 1 {
		return false, fmt.Errorf("illegal state: found more than one user matching ID %d", id)
	}
	if ct == 0 {
		return false, nil
	}
	passwd, err := hashPassword(newPassword)
	if err != nil {
		return false, err
	}
	// 🚨 SECURITY: set the new password and clear the reset code and expiry so the same code can't be reused
	if _, err := globalDB.ExecContext(ctx, "UPDATE users SET passwd_reset_code=NULL, passwd_reset_time=NULL, passwd=$1 WHERE id=$2", passwd, id); err != nil {
		return false, err
	}
	return true, nil
}

func hashPassword(password string) (sql.NullString, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return sql.NullString{}, err
	}
	return sql.NullString{Valid: true, String: string(hash)}, nil
}
