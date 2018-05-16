package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dlclark/regexp2"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
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
	AvatarURL        string // the new user's avatar URL, if known

	// EmailVerificationCode, if given, causes the new user's email address to be unverified until
	// they perform the email verification process and provied this code.
	EmailVerificationCode string `json:"-"` // forbid this field being set by JSON, just in case

	// EmailIsVerified is whether the email address should be considered already verified.
	//
	// 🚨 SECURITY: Only site admins are allowed to create users whose email addresses are initially
	// verified (i.e., with EmailVerificationCode == "" and ExternalProvider == "").
	EmailIsVerified bool `json:"-"` // forbid this field being set by JSON, just in case

	// FailIfNotInitialUser causes the (users).Create call to return an error and not create the
	// user if at least one of the following is true: (1) the site has already been initialized or
	// (2) any other user account already exists.
	FailIfNotInitialUser bool `json:"-"` // forbid this field being set by JSON, just in case
}

// Create creates a new user in the database. The provider specifies what identity providers was responsible for authenticating
// the user:
// - If the provider is empty, the user is a builtin user with no external auth account associated
// - If the provider is something else, the user was authenticated by an SSO provider
//
// Builtin users must also specify a password and email upon creation.
//
// CREATION OF SITE ADMINS
//
// The new user is made to be a site admin if the following are both true: (1) this user would be
// the first and only user on the server, and (2) the site has not yet been initialized. Otherwise,
// the user is created as a normal (non-site-admin) user. After the call, the site is marked as
// having been initialized (so that no subsequent (users).Create calls will yield a site
// admin). This is used to create the initial site admin user during site initialization.
//
// It's implemented as part of the (users).Create call instead of relying on the caller to do it in
// order to avoid a race condition where multiple initial site admins could be created or zero site
// admins could be created.
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
	if info.ExternalID == "" && (info.Password == "" || info.Email == "") {
		return nil, errors.New("no password or email provided for new builtin user")
	}
	if info.ExternalID == "" && info.EmailVerificationCode == "" && !info.EmailIsVerified {
		return nil, errors.New("no email verification code provided for new builtin user")
	}
	if info.ExternalID != "" && (info.Password != "" || info.EmailVerificationCode != "") {
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

	var avatarURL *string
	if info.AvatarURL != "" {
		avatarURL = &info.AvatarURL
	}

	dbExternalID := sql.NullString{String: info.ExternalID}
	dbExternalID.Valid = info.ExternalID != ""

	dbExternalProvider := sql.NullString{String: info.ExternalProvider}
	dbExternalProvider.Valid = info.ExternalProvider != ""

	dbEmailCode := sql.NullString{String: info.EmailVerificationCode}
	dbEmailCode.Valid = info.EmailVerificationCode != ""

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

	// Creating the initial site admin user is equivalent to initializing the
	// site. ensureInitialized runs in the transaction, so we are guaranteed that the user account
	// creation and site initialization operations occur atomically (to guarantee to the legitimate
	// site admin that if they successfully initialize the server, then no attacker's account could
	// have been created as a site admin).
	alreadyInitialized, err := (&siteConfig{}).ensureInitialized(ctx, tx)
	if err != nil {
		return nil, err
	}
	if alreadyInitialized && info.FailIfNotInitialUser {
		return nil, errCannotCreateUser{"site_already_initialized"}
	}

	var siteAdmin bool
	err = tx.QueryRowContext(
		ctx,
		"INSERT INTO users(external_id, username, display_name, avatar_url, external_provider, created_at, updated_at, passwd, site_admin) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9 AND NOT EXISTS(SELECT * FROM users)) RETURNING id, site_admin",
		dbExternalID, info.Username, info.DisplayName, avatarURL, dbExternalProvider, createdAt, updatedAt, passwd, !alreadyInitialized).Scan(&id, &siteAdmin)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Constraint {
			case "users_username":
				return nil, errCannotCreateUser{errorCodeUsernameExists}
			case "users_external_id":
				return nil, errCannotCreateUser{"err_external_id_exists"}
			}
		}
		return nil, err
	}
	if info.FailIfNotInitialUser && !siteAdmin {
		// Refuse to make the user the initial site admin if there are other existing users.
		return nil, errCannotCreateUser{"initial_site_admin_must_be_first_user"}
	}

	if info.Email != "" {
		var err error
		if info.EmailIsVerified {
			_, err = tx.ExecContext(ctx, "INSERT INTO user_emails(user_id, email, verified_at) VALUES ($1, $2, now())", id, info.Email)
		} else {
			_, err = tx.ExecContext(ctx, "INSERT INTO user_emails(user_id, email, verification_code) VALUES ($1, $2, $3)", id, info.Email, info.EmailVerificationCode)
		}
		if err != nil {
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
		AvatarURL:        info.AvatarURL,
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

// UserUpdate describes user fields to update.
type UserUpdate struct {
	Username string // update the Username to this value (if non-zero)

	// For the following fields:
	//
	// - If nil, the value in the DB is unchanged.
	// - If pointer to "" (empty string), the value in the DB is set to null.
	// - If pointer to a non-empty string, the value in the DB is set to the string.
	DisplayName, AvatarURL *string
}

// Update updates a user's profile information.
func (u *users) Update(ctx context.Context, id int32, update UserUpdate) error {
	if Mocks.Users.Update != nil {
		return Mocks.Users.Update(id, update)
	}

	fieldUpdates := []*sqlf.Query{
		sqlf.Sprintf("updated_at=now()"), // always update updated_at timestamp
	}
	if update.Username != "" {
		fieldUpdates = append(fieldUpdates, sqlf.Sprintf("username=%s", update.Username))
	}
	strOrNil := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}
	if update.DisplayName != nil {
		fieldUpdates = append(fieldUpdates, sqlf.Sprintf("display_name=%s", strOrNil(*update.DisplayName)))
	}
	if update.AvatarURL != nil {
		fieldUpdates = append(fieldUpdates, sqlf.Sprintf("avatar_url=%s", strOrNil(*update.AvatarURL)))
	}
	query := sqlf.Sprintf("UPDATE users SET %s WHERE id=%d", sqlf.Join(fieldUpdates, ", "), id)
	res, err := globalDB.ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Constraint == "users_username" {
			return errors.New("username already exists")
		}
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return userNotFoundErr{args: []interface{}{id}}
	}
	return nil
}

func (u *users) Delete(ctx context.Context, id int32) error {
	// Wrap in transaction because we delete from multiple tables.
	tx, err := globalDB.BeginTx(ctx, nil)
	if err != nil {
		return err
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

	res, err := tx.ExecContext(ctx, "UPDATE users SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL", id)
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

	if _, err := tx.ExecContext(ctx, "UPDATE access_tokens SET deleted_at=now() WHERE subject_user_id=$1 OR creator_user_id=$1", id); err != nil {
		return err
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

// GetByVerifiedEmail returns the user (if any) with the specified verified email address. If a user
// has a matching *unverified* email address, they will not be returned by this method. At most one
// user may have any given verified email address.
func (u *users) GetByVerifiedEmail(ctx context.Context, email string) (*types.User, error) {
	return u.getOneBySQL(ctx, "WHERE id=(SELECT user_id FROM user_emails WHERE email=$1 AND verified_at IS NOT NULL) AND deleted_at IS NULL LIMIT 1", email)
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
	// UserIDs specifies a list of user IDs to include.
	UserIDs []int32

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

	if len(opt.UserIDs) > 0 {
		items := []*sqlf.Query{}
		for _, id := range opt.UserIDs {
			items = append(items, sqlf.Sprintf("%d", id))
		}
		conds = append(conds, sqlf.Sprintf("u.id IN (%s)", sqlf.Join(items, ",")))
	}

	q := sqlf.Sprintf("WHERE %s ORDER BY id ASC %s", sqlf.Join(conds, "AND"), opt.LimitOffset.SQL())
	return u.getBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
}

func (*users) listSQL(opt UsersListOptions) (conds []*sqlf.Query) {
	conds = []*sqlf.Query{sqlf.Sprintf("TRUE")}
	conds = append(conds, sqlf.Sprintf("deleted_at IS NULL"))
	if opt.Query != "" {
		query := "%" + opt.Query + "%"
		conds = append(conds, sqlf.Sprintf("(username ILIKE %s OR display_name ILIKE %s)", query, query))
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
		var displayName, dbExternalID, dbExternalProvider, avatarURL sql.NullString
		err := rows.Scan(&u.ID, &dbExternalID, &u.Username, &displayName, &dbExternalProvider, &avatarURL, &u.CreatedAt, &u.UpdatedAt, &u.SiteAdmin)
		if err != nil {
			return nil, err
		}
		u.DisplayName = displayName.String
		u.AvatarURL = avatarURL.String
		if dbExternalID.Valid {
			u.ExternalID = &dbExternalID.String
		}
		if dbExternalProvider.Valid {
			u.ExternalProvider = dbExternalProvider.String
		}
		users = append(users, &u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
