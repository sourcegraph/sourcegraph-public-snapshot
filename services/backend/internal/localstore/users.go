package localstore

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/gorp.v1"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
)

func init() {
	AppSchema.Map.AddTableWithName(dbUser{}, "users").SetKeys(true, "UID")
	AppSchema.CreateSQL = append(AppSchema.CreateSQL,
		"ALTER TABLE users ALTER COLUMN login TYPE citext",
		"CREATE UNIQUE INDEX users_login ON users(login)",
		`ALTER TABLE users ALTER COLUMN registered_at TYPE timestamp with time zone USING registered_at::timestamp with time zone;`,
		`CREATE INDEX users_login_ci ON users((lower(login)) text_pattern_ops);`,
		`ALTER TABLE users ALTER COLUMN betas TYPE text ARRAY USING betas::text[]`,
	)
}

// dbUser DB-maps a sourcegraph.User object.
type dbUser struct {
	UID            int
	Login          string
	Name           string
	IsOrganization bool
	AvatarURL      string `db:"avatar_url"`
	Location       string
	Company        string
	HomepageURL    string `db:"homepage_url"`
	Disabled       bool   `db:"disabled"`
	Write          bool
	Admin          bool
	Betas          *dbutil.StringSlice
	BetaRegistered bool       `db:"beta_registered"`
	RegisteredAt   *time.Time `db:"registered_at"`
}

func (u *dbUser) toUser() *sourcegraph.User {
	var betas []string
	if u.Betas != nil {
		betas = u.Betas.Slice
	}
	return &sourcegraph.User{
		UID:            int32(u.UID),
		Login:          u.Login,
		Name:           u.Name,
		IsOrganization: u.IsOrganization,
		AvatarURL:      u.AvatarURL,
		Location:       u.Location,
		Company:        u.Company,
		HomepageURL:    u.HomepageURL,
		Disabled:       u.Disabled,
		Write:          u.Write,
		Admin:          u.Admin,
		Betas:          betas,
		BetaRegistered: u.BetaRegistered,
		RegisteredAt:   ts(u.RegisteredAt),
	}
}

func (u *dbUser) fromUser(u2 *sourcegraph.User) {
	u.UID = int(u2.UID)
	u.Login = u2.Login
	u.Name = u2.Name
	u.IsOrganization = u2.IsOrganization
	u.AvatarURL = u2.AvatarURL
	u.Location = u2.Location
	u.Company = u2.Company
	u.HomepageURL = u2.HomepageURL
	u.Disabled = u2.Disabled
	u.Write = u2.Write
	u.Admin = u2.Admin
	if len(u2.Betas) > 0 {
		u.Betas = &dbutil.StringSlice{Slice: u2.Betas}
	}
	u.BetaRegistered = u2.BetaRegistered
	u.RegisteredAt = tm(u2.RegisteredAt)
}

func toUsers(us []*dbUser) []*sourcegraph.User {
	u2s := make([]*sourcegraph.User, len(us))
	for i, u := range us {
		u2s[i] = u.toUser()
	}
	return u2s
}

// users is a DB-backed implementation of the Users store.
type users struct{}

var _ store.Users = (*users)(nil)

func (s *users) Get(ctx context.Context, userSpec sourcegraph.UserSpec) (*sourcegraph.User, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Users.Get", nil); err != nil {
		return nil, err
	}
	var user *sourcegraph.User
	var err error
	if userSpec.UID != 0 && userSpec.Login != "" {
		user, err = s.getBySQL(ctx, "uid=$1 AND login=$2", userSpec.UID, userSpec.Login)
		if err == sql.ErrNoRows {
			err = &store.UserNotFoundError{UID: int(userSpec.UID), Login: userSpec.Login}
		}
	} else if userSpec.UID != 0 {
		user, err = s.getByUID(ctx, int(userSpec.UID))
	} else if userSpec.Login != "" {
		user, err = s.getByLogin(ctx, userSpec.Login)
	} else {
		return nil, &store.UserNotFoundError{}
	}
	return user, err
}

// getByUID returns the user with the given uid, if such a user
// exists in the database.
func (s *users) getByUID(ctx context.Context, uid int) (*sourcegraph.User, error) {
	user, err := s.getBySQL(ctx, "uid=$1", uid)
	if err == sql.ErrNoRows {
		err = &store.UserNotFoundError{UID: uid}
	}
	return user, err
}

// getByLogin returns the user with the given login, if such a user
// exists in the database.
func (s *users) getByLogin(ctx context.Context, login string) (*sourcegraph.User, error) {
	user, err := s.getBySQL(ctx, "login=$1", login)
	if err == sql.ErrNoRows {
		err = &store.UserNotFoundError{Login: login}
	}
	return user, err
}

// getBySQL returns a user matching the SQL query (if any exists). A
// "LIMIT 1" clause is appended to the query before it is executed.
func (s *users) getBySQL(ctx context.Context, query string, args ...interface{}) (*sourcegraph.User, error) {
	var user dbUser
	if err := appDBH(ctx).SelectOne(&user, "SELECT * FROM users WHERE ("+query+") LIMIT 1", args...); err != nil {
		return nil, err
	}
	return user.toUser(), nil
}

var okUsersSorts = map[string]struct{}{
	"login":        struct{}{},
	"lower(login)": struct{}{},
	"uid":          struct{}{},
}

func (s *users) List(ctx context.Context, opt *sourcegraph.UsersListOptions) ([]*sourcegraph.User, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Users.List", nil); err != nil {
		return nil, err
	}
	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}
	sql := fmt.Sprintf(`FROM users WHERE NOT disabled`)
	if opt.Query != "" {
		sql += " AND (LOWER(login)=" + arg(strings.ToLower(opt.Query)) + " OR LOWER(login) LIKE " + arg(strings.ToLower(opt.Query)+"%") + ")"
	}

	if opt.UIDs != nil && len(opt.UIDs) > 0 {
		uidBindVars := make([]string, len(opt.UIDs))
		for i, uid := range opt.UIDs {
			uidBindVars[i] = arg(uid)
		}
		sql += " AND uid in (" + strings.Join(uidBindVars, ",") + ")"
	}

	for _, beta := range opt.AllBetas {
		bindVar := arg(beta)
		sql += " AND " + bindVar + " = ANY(betas)"
	}
	if opt.RegisteredBeta {
		// Filter by users who have registered for beta access.
		sql += " AND beta_registered=true "
	}
	if opt.HaveBeta && len(opt.AllBetas) == 0 {
		// Filter by users who have access to at least one beta. Note that
		// len(opt.AllBetas) > 0 fulfils this requirement.
		sql += " AND array_length(betas, 1) > 0 "
	}

	sort := opt.Sort
	direction := opt.Direction
	if sort == "" {
		sort = "lower(login)"
	}
	if _, ok := okUsersSorts[sort]; !ok {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid sort: "+sort)
	}

	if direction == "" {
		if sort == "uid" || sort == "lower(login)" {
			direction = "asc"
		} else {
			direction = "desc"
		}
	}
	if direction != "asc" && direction != "desc" {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid direction: "+direction)
	}

	sql += fmt.Sprintf(" ORDER BY %s %s", sort, strings.ToUpper(direction))
	sql += fmt.Sprintf(" LIMIT %s OFFSET %s", arg(opt.PerPageOrDefault()), arg(opt.Offset()))

	sql = "SELECT * " + sql
	var users []*dbUser
	if _, err := appDBH(ctx).Select(&users, sql, args...); err != nil {
		return nil, err
	}
	return toUsers(users), nil
}

func (s *users) GetUIDByGitHubID(ctx context.Context, githubUID int) (int32, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Users.GetUIDByGitHubID", nil); err != nil {
		return 0, err
	}
	uid, err := appDBH(ctx).SelectInt(`SELECT "user" FROM ext_auth_token WHERE host='github.com' AND (NOT disabled) AND ext_uid=$1;`, githubUID)
	if err == sql.ErrNoRows || uid == 0 {
		err = grpc.Errorf(codes.NotFound, "no external auth token for github user %d", githubUID)
	}
	return int32(uid), err
}
