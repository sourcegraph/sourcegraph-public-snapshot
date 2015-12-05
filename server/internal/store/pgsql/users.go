package pgsql

import (
	"fmt"
	"strings"
	"time"

	"github.com/sqs/modl"
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

func init() {
	Schema.Map.AddTableWithName(dbUser{}, "users").SetKeys(true, "UID")
	Schema.CreateSQL = append(Schema.CreateSQL,
		"ALTER TABLE users ALTER COLUMN login TYPE citext",
		"CREATE UNIQUE INDEX users_login ON users(login)",
		`ALTER TABLE users ALTER COLUMN registered_at TYPE timestamp with time zone USING registered_at::timestamp with time zone;`,
		`CREATE INDEX users_login_ci ON users((lower(login)) text_pattern_ops);`,
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
	RegisteredAt   *time.Time `db:"registered_at"`
}

func (u *dbUser) toUser() *sourcegraph.User {
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
	u.RegisteredAt = tm(u2.RegisteredAt)
}

func toUsers(us []*dbUser) []*sourcegraph.User {
	u2s := make([]*sourcegraph.User, len(us))
	for i, u := range us {
		u2s[i] = u.toUser()
	}
	return u2s
}

// Users is a DB-backed implementation of the Users store.
type Users struct{}

var _ store.Users = (*Users)(nil)

func (s *Users) Get(ctx context.Context, userSpec sourcegraph.UserSpec) (*sourcegraph.User, error) {
	var user *sourcegraph.User
	var err error
	if userSpec.UID != 0 && userSpec.Login != "" {
		user, err = s.getBySQL(ctx, "uid=$1 AND login=$2", userSpec.UID, userSpec.Login)
	} else if userSpec.UID != 0 {
		user, err = s.getBySQL(ctx, "uid=$1", userSpec.UID)
	} else if userSpec.Login != "" {
		user, err = s.getBySQL(ctx, "login=$1", userSpec.Login)
	} else {
		return nil, &store.UserNotFoundError{}
	}
	return user, err
}

// getByUID returns the user with the given uid, if such a user
// exists in the database.
func (s *Users) getByUID(ctx context.Context, uid int) (*sourcegraph.User, error) {
	return s.getBySQL(ctx, "uid=$1", uid)
}

// getByLogin returns the user with the given login, if such a user
// exists in the database.
func (s *Users) getByLogin(ctx context.Context, login string) (*sourcegraph.User, error) {
	return s.getBySQL(ctx, "login=$1", login)
}

// getBySQL returns a user matching the SQL query (if any exists). A
// "LIMIT 1" clause is appended to the query before it is executed.
func (s *Users) getBySQL(ctx context.Context, sql string, args ...interface{}) (*sourcegraph.User, error) {
	var users []*dbUser
	err := dbh(ctx).Select(&users, "SELECT * FROM users WHERE ("+sql+") LIMIT 1", args...)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, &store.UserNotFoundError{Login: "(from args)"} // can't nicely serialize args
	}
	return users[0].toUser(), nil
}

var okUsersSorts = map[string]struct{}{
	"login":        struct{}{},
	"lower(login)": struct{}{},
	"uid":          struct{}{},
}

func (s *Users) List(ctx context.Context, opt *sourcegraph.UsersListOptions) ([]*sourcegraph.User, error) {
	var args []interface{}
	arg := func(a interface{}) string {
		v := modl.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}
	sql := fmt.Sprintf(`FROM users WHERE NOT disabled`)
	if opt.Query != "" {
		sql += " AND (LOWER(login)=" + arg(strings.ToLower(opt.Query)) + " OR LOWER(login) LIKE " + arg(strings.ToLower(opt.Query)+"%") + ")"
	}

	sort := opt.Sort
	direction := opt.Direction
	if sort == "" {
		sort = "lower(login)"
	}
	if _, ok := okUsersSorts[sort]; !ok {
		return nil, &sourcegraph.InvalidOptionsError{Reason: "invalid sort: " + sort}
	}

	if direction == "" {
		if sort == "uid" || sort == "lower(login)" {
			direction = "asc"
		} else {
			direction = "desc"
		}
	}
	if direction != "asc" && direction != "desc" {
		return nil, &sourcegraph.InvalidOptionsError{Reason: "invalid direction: " + direction}
	}

	sql += fmt.Sprintf(" ORDER BY %s %s", sort, strings.ToUpper(direction))
	sql += fmt.Sprintf(" LIMIT %s OFFSET %s", arg(opt.PerPageOrDefault()), arg(opt.Offset()))

	sql = "SELECT * " + sql
	var users []*dbUser
	if err := dbh(ctx).Select(&users, sql, args...); err != nil {
		return nil, err
	}
	return toUsers(users), nil
}

func (s *Users) Count(ctx context.Context) (int32, error) {
	sql := "SELECT count(*) FROM users WHERE NOT disabled;"
	var count []int
	if err := dbh(ctx).Select(&count, sql); err != nil || len(count) == 0 {
		return 0, err
	}
	return int32(count[0]), nil
}
