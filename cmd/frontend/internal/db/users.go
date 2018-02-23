package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/dlclark/regexp2"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	"golang.org/x/crypto/bcrypt"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/lib/pq"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
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

// userNotFoundErr is the error that is returned when a user is not found.
type userNotFoundErr struct {
	args []interface{}
}

func (err userNotFoundErr) Error() string {
	return fmt.Sprintf("user not found: %v", err.args)
}

func (err userNotFoundErr) NotFound() bool {
	return true
}

// errCannotCreateUser is the error that is returned when
// a user cannot be added to the DB due to a constraint.
type errCannotCreateUser struct {
	code string
}

const (
	errorCodeUsernameExists = "err_username_exists"
	errorCodeEmailExists    = "err_email_exists"
)

func (err errCannotCreateUser) Error() string {
	return fmt.Sprintf("cannot create user: %v", err.code)
}

func (err errCannotCreateUser) Code() string {
	return err.code
}

// IsUsernameExists reports whether err is an error indicating that the intended username exists.
func IsUsernameExists(err error) bool {
	e, ok := err.(errCannotCreateUser)
	return ok && e.code == errorCodeUsernameExists
}

// IsEmailExists reports whether err is an error indicating that the intended username exists.
func IsEmailExists(err error) bool {
	e, ok := err.(errCannotCreateUser)
	return ok && e.code == errorCodeEmailExists
}

// NewUser describes a new to-be-created user.
type NewUser struct {
	ExternalID       string
	Email            string
	Username         string
	DisplayName      string
	ExternalProvider string
	Password         string
	EmailCode        string

	// InitialSiteAdminOrFail indicates that the newly created user should be made the site's initial site admin if
	// the following are all true: (1) this user would be the first and only user on the server, and (2) the site
	// has not yet been initialized. Otherwise, the (users).Create call fails and no new user is created.
	//
	// If the call with InitialSiteAdminOrFail == true succeeds, it marks the site as being initialized in the DB
	// so that subsequent calls with InitialSiteAdminOrFail == true will fail.
	//
	// It is used to create the initial site admin user during site initialization, and it's implemented as part of
	// the (users).Create call to avoid a race condition where multiple initial site admins could be created or
	// zero site admins could be created.
	InitialSiteAdminOrFail bool `json:"-"` // forbid this field being set by JSON, just in case
}

// Create creates a new user in the database. The provider specifies what identity providers was responsible for authenticating
// the user:
// - If the provider is empty, the user is a builtin user with no external auth account associated
// - If the provider is something else, the user was authenticated by an SSO provider
//
// Builtin users must also specify a password and email verification code upon creation. When the user's email is
// verified, the email verification code is set to null in the DB. All other users have a null password and email verification code.
func (*users) Create(ctx context.Context, info NewUser) (newUser *types.User, err error) {
	if Mocks.Users.Create != nil {
		return Mocks.Users.Create(ctx, info)
	}

	if info.ExternalID != "" && info.ExternalProvider == "" {
		return nil, errors.New("external ID is set but external provider is empty")
	}
	if info.ExternalProvider != "" && info.ExternalID == "" {
		return nil, errors.New("external provider is set but external ID is empty")
	}
	if info.ExternalID == "" && (info.Password == "" || info.EmailCode == "") {
		return nil, errors.New("no password or email code provided for new builtin user")
	}
	if info.ExternalID != "" && (info.Password != "" || info.EmailCode != "") {
		return nil, errors.New("password and/or email verification code provided for external user")
	}

	createdAt := time.Now()
	updatedAt := createdAt
	var id int32

	var passwd sql.NullString
	if info.Password == "" {
		passwd = sql.NullString{Valid: false}
	} else {
		// Compute hash of password
		passwd, err = hashPassword(info.Password)
		if err != nil {
			return nil, err
		}
	}

	dbExternalID := sql.NullString{String: info.ExternalID}
	dbExternalID.Valid = info.ExternalID != ""

	dbExternalProvider := sql.NullString{String: info.ExternalProvider}
	dbExternalProvider.Valid = info.ExternalProvider != ""

	dbEmailCode := sql.NullString{String: info.EmailCode}
	dbEmailCode.Valid = info.EmailCode != ""

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

	if info.InitialSiteAdminOrFail {
		// The "SELECT ... FOR UPDATE" prevents a race condition where two calls, each in their own transaction,
		// would see this initialized value as false.
		var siteInitialized bool
		if err := tx.QueryRowContext(ctx, `SELECT initialized FROM site_config FOR UPDATE LIMIT 1`).Scan(&siteInitialized); err != nil {
			return nil, err
		}
		if siteInitialized {
			return nil, errCannotCreateUser{"site_already_initialized"}
		}

		// Creating the initial site admin user is equivalent to initializing the site. This prevents other initial
		// site admin users from being created (to prevent a race condition where an attacker could create a site
		// admin account simultaneously with the real site admin).
		if _, err := tx.ExecContext(ctx, `UPDATE site_config SET initialized=true`); err != nil {
			return nil, err
		}
	}

	var siteAdmin bool
	err = tx.QueryRowContext(
		ctx,
		"INSERT INTO users(external_id, username, display_name, external_provider, created_at, updated_at, passwd, site_admin) VALUES($1, $2, $3, $4, $5, $6, $7, $8 AND NOT EXISTS(SELECT * FROM users)) RETURNING id, site_admin",
		dbExternalID, info.Username, info.DisplayName, dbExternalProvider, createdAt, updatedAt, passwd, info.InitialSiteAdminOrFail).Scan(&id, &siteAdmin)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Constraint {
			case "users_username_key":
				return nil, errCannotCreateUser{errorCodeUsernameExists}
			case "users_external_id":
				return nil, errCannotCreateUser{"err_external_id_exists"}
			}
		}
		return nil, err
	}
	if info.InitialSiteAdminOrFail && !siteAdmin {
		// Refuse to make the user the initial site admin if there are other existing users.
		return nil, errCannotCreateUser{"initial_site_admin_must_be_first_user"}
	}

	if info.Email != "" {
		if _, err := tx.ExecContext(ctx, "INSERT INTO user_emails(user_id, email, verification_code) VALUES ($1, $2, $3)",
			id, info.Email, info.EmailCode,
		); err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Constraint {
				case "user_emails_email_key":
					return nil, errCannotCreateUser{errorCodeEmailExists}
				}
			}
			return nil, err
		}
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

	return &types.User{
		ID:               id,
		ExternalID:       &info.ExternalID,
		Username:         info.Username,
		DisplayName:      info.DisplayName,
		ExternalProvider: info.ExternalProvider,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		SiteAdmin:        siteAdmin,
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

func (u *users) Update(ctx context.Context, id int32, username *string, displayName *string, avatarURL *string) (*types.User, error) {
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

func (u *users) Delete(ctx context.Context, id int32) error {
	res, err := globalDB.ExecContext(ctx, "UPDATE users SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return userNotFoundErr{args: []interface{}{id}}
	}
	return nil
}

func (u *users) SetIsSiteAdmin(ctx context.Context, id int32, isSiteAdmin bool) error {
	_, err := globalDB.ExecContext(ctx, "UPDATE users SET site_admin=$1 WHERE id=$2", isSiteAdmin, id)
	return err
}

// CheckAndDecrementInviteQuota should be called before the user (identified
// by userID) is allowed to invite any other user. If ok is false, then the
// user is not allowed to invite any other user (either because they've
// invited too many users, or some other error occurred). If the user has
// quota remaining, their quota is decremented and ok is true.
func (u *users) CheckAndDecrementInviteQuota(ctx context.Context, userID int32) (ok bool, err error) {
	var quotaRemaining int32
	sqlQuery := `
	UPDATE users SET invite_quota=(invite_quota - 1)
	WHERE users.id=$1 AND invite_quota>0 AND deleted_at IS NULL
	RETURNING invite_quota`
	row := globalDB.QueryRowContext(ctx, sqlQuery, userID)
	if err := row.Scan(&quotaRemaining); err == sql.ErrNoRows {
		// It's possible that some other problem occurred, such as the user being deleted,
		// but treat that as a quota exceeded error, too.
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil // the user has remaining quota to send invites
}

func (u *users) GetByID(ctx context.Context, id int32) (*types.User, error) {
	if Mocks.Users.GetByID != nil {
		return Mocks.Users.GetByID(ctx, id)
	}
	return u.getOneBySQL(ctx, "WHERE id=$1 AND deleted_at IS NULL LIMIT 1", id)
}

// GetByExternalID gets the user (if any) from the database that is associated with an external
// user account, based on the given provider and ID on the provider.
func (u *users) GetByExternalID(ctx context.Context, provider, id string) (*types.User, error) {
	if provider == "" || id == "" {
		panic(fmt.Sprintf("GetByExternalID: both provider (%q) and id (%q) must be nonempty", provider, id))
	}
	if Mocks.Users.GetByExternalID != nil {
		return Mocks.Users.GetByExternalID(ctx, provider, id)
	}
	return u.getOneBySQL(ctx, "WHERE external_provider=$1 AND external_id=$2 AND deleted_at IS NULL LIMIT 1", provider, id)
}

func (u *users) GetByEmail(ctx context.Context, email string) (*types.User, error) {
	return u.getOneBySQL(ctx, "WHERE id=(SELECT user_id FROM user_emails WHERE email=$1) AND deleted_at IS NULL LIMIT 1", email)
}

func (u *users) GetByUsername(ctx context.Context, username string) (*types.User, error) {
	if Mocks.Users.GetByUsername != nil {
		return Mocks.Users.GetByUsername(ctx, username)
	}

	return u.getOneBySQL(ctx, "WHERE username=$1 AND deleted_at IS NULL LIMIT 1", username)
}

var ErrNoCurrentUser = errors.New("no current user")

func (u *users) GetByCurrentAuthUser(ctx context.Context) (*types.User, error) {
	if Mocks.Users.GetByCurrentAuthUser != nil {
		return Mocks.Users.GetByCurrentAuthUser(ctx)
	}

	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, ErrNoCurrentUser
	}

	return u.getOneBySQL(ctx, "WHERE id=$1 AND deleted_at IS NULL LIMIT 1", actor.UID)
}

func (u *users) Count(ctx context.Context, opt UsersListOptions) (int, error) {
	if Mocks.Users.Count != nil {
		return Mocks.Users.Count(ctx, opt)
	}

	conds := u.listSQL(opt)
	q := sqlf.Sprintf("SELECT COUNT(*) FROM users WHERE %s", sqlf.Join(conds, "AND"))

	var count int
	if err := globalDB.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// ListByOrg returns users for a given org. It can also query a list of specific
// users by either user IDs or usernames.
func (u *users) ListByOrg(ctx context.Context, orgID int32, userIDs []int32, usernames []string) ([]*types.User, error) {
	if Mocks.Users.ListByOrg != nil {
		return Mocks.Users.ListByOrg(ctx, orgID, userIDs, usernames)
	}
	conds := []*sqlf.Query{}
	filters := []*sqlf.Query{}
	if len(userIDs) > 0 {
		items := []*sqlf.Query{}
		for _, id := range userIDs {
			items = append(items, sqlf.Sprintf("%d", id))
		}
		filters = append(filters, sqlf.Sprintf("u.id IN (%s)", sqlf.Join(items, ",")))
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
	q := sqlf.Sprintf("JOIN org_members ON (org_members.user_id = u.id) WHERE %s", sqlf.Join(conds, "AND"))
	return u.getBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
}

// UsersListOptions specifies the options for listing users.
type UsersListOptions struct {
	// Query specifies a search query for users.
	Query string

	*LimitOffset
}

func (u *users) List(ctx context.Context, opt *UsersListOptions) ([]*types.User, error) {
	if Mocks.Users.List != nil {
		return Mocks.Users.List(ctx, opt)
	}

	if opt == nil {
		opt = &UsersListOptions{}
	}
	conds := u.listSQL(*opt)

	q := sqlf.Sprintf("WHERE %s ORDER BY id ASC %s", sqlf.Join(conds, "AND"), opt.LimitOffset.SQL())
	return u.getBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
}

func (*users) listSQL(opt UsersListOptions) (conds []*sqlf.Query) {
	conds = []*sqlf.Query{sqlf.Sprintf("TRUE")}
	conds = append(conds, sqlf.Sprintf("deleted_at IS NULL"))
	if opt.Query != "" {
		query := "%" + opt.Query + "%"
		conds = append(conds, sqlf.Sprintf("username ILIKE %s OR display_name ILIKE %s", query, query))
	}
	return conds
}

func (u *users) getOneBySQL(ctx context.Context, query string, args ...interface{}) (*types.User, error) {
	users, err := u.getBySQL(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(users) != 1 {
		return nil, userNotFoundErr{args}
	}
	return users[0], nil
}

// getBySQL returns users matching the SQL query, if any exist.
func (*users) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*types.User, error) {
	rows, err := globalDB.QueryContext(ctx, "SELECT u.id, u.external_id, u.username, u.display_name, u.external_provider, u.avatar_url, u.created_at, u.updated_at, u.site_admin FROM users u "+query, args...)
	if err != nil {
		return nil, err
	}

	users := []*types.User{}
	defer rows.Close()
	for rows.Next() {
		var u types.User
		var dbExternalID, dbExternalProvider, avatarURL sql.NullString
		err := rows.Scan(&u.ID, &dbExternalID, &u.Username, &u.DisplayName, &dbExternalProvider, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt, &u.SiteAdmin)
		if err != nil {
			return nil, err
		}
		if dbExternalID.Valid {
			u.ExternalID = &dbExternalID.String
		}
		if dbExternalProvider.Valid {
			u.ExternalProvider = dbExternalProvider.String
		}
		if avatarURL.Valid {
			u.AvatarURL = &avatarURL.String
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
	if err := globalDB.QueryRowContext(ctx, "SELECT passwd FROM users WHERE deleted_at IS NULL AND id=$1", id).Scan(&passwd); err != nil {
		return false, err
	}
	if !passwd.Valid {
		return false, nil
	}
	return validPassword(passwd.String, password), nil
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
	r := globalDB.QueryRowContext(ctx, "SELECT count(*) FROM users WHERE id=$1 AND deleted_at IS NULL AND passwd_reset_code=$2 AND passwd_reset_time + interval '4 hours' > now()", id, resetCode)
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
	// 🚨 SECURITY: set the new password and clear the reset code and expiry so the same code can't be reused.
	if _, err := globalDB.ExecContext(ctx, "UPDATE users SET passwd_reset_code=NULL, passwd_reset_time=NULL, passwd=$1 WHERE id=$2", passwd, id); err != nil {
		return false, err
	}
	return true, nil
}

// UpdatePassword updates a user's password given the current password.
func (u *users) UpdatePassword(ctx context.Context, id int32, oldPassword, newPassword string) error {
	// 🚨 SECURITY: No empty passwords.
	if oldPassword == "" {
		return errors.New("old password was empty")
	}
	if newPassword == "" {
		return errors.New("new password was empty")
	}
	// 🚨 SECURITY: Make sure the caller provided the correct old password.
	if ok, err := u.IsPassword(ctx, id, oldPassword); err != nil {
		return err
	} else if !ok {
		return errors.New("wrong old password")
	}

	passwd, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	// 🚨 SECURITY: Set the new password
	if _, err := globalDB.ExecContext(ctx, "UPDATE users SET passwd_reset_code=NULL, passwd_reset_time=NULL, passwd=$1 WHERE id=$2", passwd, id); err != nil {
		return err
	}
	return nil
}

// RandomizePasswordAndClearPasswordResetRateLimit overwrites a user's password with a hard-to-guess
// random password and clears the password reset rate limit. It is intended to be used by site admins,
// who can subsequently generate a new password reset code for the user (in case the user has locked
// themselves out, or in case the site admin wants to initiate a password reset).
//
// A randomized password is used (instead of an empty password) to avoid bugs where an empty password
// is considered to be no password. The random password is expected to be irretrievable.
func (u *users) RandomizePasswordAndClearPasswordResetRateLimit(ctx context.Context, id int32) error {
	passwd, err := hashPassword(randstring.NewLen(36))
	if err != nil {
		return err
	}
	// 🚨 SECURITY: Set the new random password and clear the reset code/expiry, so the old code
	// can't be reused, and so a new valid reset code can be generated afterward.
	_, err = globalDB.ExecContext(ctx, "UPDATE users SET passwd_reset_code=NULL, passwd_reset_time=NULL, passwd=$1 WHERE id=$2", passwd, id)
	return err
}

// mockHashPassword if non-nil is used instead of hashPassword. This is useful
// when running tests since we can use a faster implementation.
var mockHashPassword func(password string) (sql.NullString, error)
var mockValidPassword func(hash, password string) bool

func hashPassword(password string) (sql.NullString, error) {
	if mockHashPassword != nil {
		return mockHashPassword(password)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return sql.NullString{}, err
	}
	return sql.NullString{Valid: true, String: string(hash)}, nil
}

func validPassword(hash, password string) bool {
	if mockValidPassword != nil {
		return mockValidPassword(hash, password)
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
