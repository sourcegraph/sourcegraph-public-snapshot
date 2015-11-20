package pgsql

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/sqs/modl"
	"golang.org/x/net/context"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

func init() {
	Schema.Map.AddTableWithName(dbRepo{}, "repo").SetKeys(true, "URI")
	Schema.CreateSQL = append(Schema.CreateSQL,
		"ALTER TABLE repo ALTER COLUMN uri TYPE citext",
		"ALTER TABLE repo ALTER COLUMN description TYPE text",
		`ALTER TABLE repo ALTER COLUMN default_branch SET NOT NULL;`,
		`ALTER TABLE repo ALTER COLUMN vcs SET NOT NULL;`,
		"CREATE INDEX repo_name ON repo(name text_pattern_ops);",

		// fast for repo searching by URI and name
		"CREATE INDEX repo_lower_uri_lower_name ON repo((lower(uri)::text) text_pattern_ops, lower(name));",
	)
}

// dbRepo DB-maps a sourcegraph.Repo object.
type dbRepo struct {
	URI           string
	Origin        string
	Name          string
	Description   string
	VCS           string
	HTTPCloneURL  string `db:"http_clone_url"`
	SSHCloneURL   string `db:"ssh_clone_url"`
	HomepageURL   string `db:"homepage_url"`
	DefaultBranch string `db:"default_branch"`
	Language      string
	Blocked       bool
	Deprecated    bool
	Fork          bool
	Mirror        bool
	Private       bool
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
	PushedAt      time.Time `db:"pushed_at"`
}

func (r *dbRepo) toRepo() *sourcegraph.Repo {
	return &sourcegraph.Repo{
		URI:           r.URI,
		Origin:        r.Origin,
		Name:          r.Name,
		Description:   r.Description,
		VCS:           r.VCS,
		HTTPCloneURL:  r.HTTPCloneURL,
		SSHCloneURL:   r.SSHCloneURL,
		HomepageURL:   r.HomepageURL,
		DefaultBranch: r.DefaultBranch,
		Language:      r.Language,
		Blocked:       r.Blocked,
		Deprecated:    r.Deprecated,
		Fork:          r.Fork,
		Mirror:        r.Mirror,
		Private:       r.Private,
		CreatedAt:     pbtypes.NewTimestamp(r.CreatedAt),
		UpdatedAt:     pbtypes.NewTimestamp(r.UpdatedAt),
		PushedAt:      pbtypes.NewTimestamp(r.PushedAt),
	}
}

func (r *dbRepo) fromRepo(r2 *sourcegraph.Repo) {
	r.URI = r2.URI
	r.Origin = r2.Origin
	r.Name = r2.Name
	r.Description = r2.Description
	r.VCS = r2.VCS
	r.HTTPCloneURL = r2.HTTPCloneURL
	r.SSHCloneURL = r2.SSHCloneURL
	r.HomepageURL = r2.HomepageURL
	r.DefaultBranch = r2.DefaultBranch
	r.Language = r2.Language
	r.Blocked = r2.Blocked
	r.Deprecated = r2.Deprecated
	r.Fork = r2.Fork
	r.Mirror = r2.Mirror
	r.Private = r2.Private
	r.CreatedAt = r2.CreatedAt.Time()
	r.UpdatedAt = r2.UpdatedAt.Time()
	r.PushedAt = r2.PushedAt.Time()
}

func toRepos(rs []*dbRepo) []*sourcegraph.Repo {
	r2s := make([]*sourcegraph.Repo, len(rs))
	for i, r := range rs {
		r2s[i] = r.toRepo()
	}
	return r2s
}

// Repos is a DB-backed implementation of the Repos store.
type Repos struct{}

var _ store.Repos = (*Repos)(nil)

func (s *Repos) Get(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
	return s.getByURI(ctx, repo)
}

func (s *Repos) getByURI(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	repo, err := s.getBySQL(ctx, "uri=$1", uri)
	if err != nil {
		if e, ok := err.(*store.RepoNotFoundError); ok {
			e.Repo = uri
		}
	}
	return repo, err
}

// getBySQL returns a repository matching the SQL query, if any
// exists. A "LIMIT 1" clause is appended to the query before it is
// executed.
func (s *Repos) getBySQL(ctx context.Context, sql string, args ...interface{}) (*sourcegraph.Repo, error) {
	var repos []*dbRepo
	err := dbh(ctx).Select(&repos, "SELECT * FROM repo WHERE ("+sql+") LIMIT 1", args...)
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 {
		return nil, &store.RepoNotFoundError{Repo: "(unknown)"} // can't nicely serialize args
	}
	return repos[0].toRepo(), nil
}

func (s *Repos) GetPerms(ctx context.Context, repo string) (*sourcegraph.RepoPermissions, error) {
	return &sourcegraph.RepoPermissions{Read: true, Write: true, Admin: true}, nil
}

func (s *Repos) List(ctx context.Context, opt *sourcegraph.RepoListOptions) ([]*sourcegraph.Repo, error) {
	if opt == nil {
		opt = &sourcegraph.RepoListOptions{}
	}

	sql, args, err := s.listSQL(opt)
	if err != nil {
		if err == errOptionsSpecifyEmptyResult {
			err = nil
		}
		return nil, err
	}

	arg := func(a interface{}) string {
		v := modl.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	// LIMIT and OFFSET
	sql += fmt.Sprintf(" LIMIT %s OFFSET %s", arg(opt.PerPageOrDefault()), arg(opt.Offset()))

	repos, err := s.query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

var errOptionsSpecifyEmptyResult = errors.New("pgsql: options specify and empty result set")

func (s *Repos) listSQL(opt *sourcegraph.RepoListOptions) (string, []interface{}, error) {
	var selectSQL, fromSQL, whereSQL, orderBySQL string

	var args []interface{}
	arg := func(a interface{}) string {
		v := modl.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	queryTerms := strings.Fields(opt.Query)
	uriQuery := strings.ToLower(strings.Join(queryTerms, "/"))

	{ // SELECT
		selectSQL = "repo.*"
	}
	{ // FROM
		fromSQL = "repo"
	}
	{ // WHERE
		var conds []string

		conds = append(conds, "(NOT blocked)")

		if opt.NoFork {
			conds = append(conds, "(NOT fork)")
		}
		if len(opt.URIs) > 0 {
			if len(opt.URIs) == 1 && strings.Contains(opt.URIs[0], ",") {
				// Workaround for https://github.com/sourcegraph/go-sourcegraph/issues/30.
				opt.URIs = strings.Split(opt.URIs[0], ",")
			}

			uriBindVars := make([]string, len(opt.URIs))
			for i, uri := range opt.URIs {
				uriBindVars[i] = arg(uri)
			}
			conds = append(conds, "uri IN ("+strings.Join(uriBindVars, ",")+")")
		}
		if opt.Name != "" {
			conds = append(conds, "lower(name)="+arg(strings.ToLower(opt.Name)))
		}
		if len(queryTerms) >= 1 {
			uriQuery = strings.ToLower(uriQuery)
			conds = append(conds, "lower(uri) LIKE "+arg("/"+uriQuery+"%")+" OR lower(uri) LIKE "+arg(uriQuery+"%/%")+" OR lower(name) LIKE "+arg(uriQuery+"%")+" OR lower(uri) = "+arg(uriQuery))
		}
		switch opt.Type {
		case "private":
			conds = append(conds, `private`)
		case "public":
			conds = append(conds, `NOT private`)
		case "", "all":
		default:
			return "", nil, &sourcegraph.InvalidOptionsError{Reason: "invalid state"}
		}
		switch opt.State {
		case "enabled", "disabled":
			sql := `EXISTS (SELECT NULL FROM repo_config rc WHERE rc.repo=repo.uri AND enabled='true')`
			if opt.State == "disabled" {
				sql = "NOT " + sql
			}
			conds = append(conds, sql)
		case "", "all":
		default:
			return "", nil, &sourcegraph.InvalidOptionsError{Reason: "invalid state"}
		}
		if opt.BuiltOnly {
			conds = append(conds, "EXISTS (SELECT NULL FROM repo_build WHERE success AND repo_build.repo=repo.uri AND NOT purged)")
		}
		if opt.Owner != "" {
			return "", nil, errOptionsSpecifyEmptyResult
		}

		if conds != nil {
			whereSQL = "(" + strings.Join(conds, ") AND (") + ")"
		} else {
			whereSQL = "true"
		}
	}

	// ORDER BY
	if uriQuery != "" {
		orderBySQL = fmt.Sprintf("(lower(name) = %s) DESC, ", arg(strings.ToLower(path.Base(uriQuery))))
	}
	sort := opt.Sort
	if sort == "" {
		sort = "uri"
	}
	sortKeyToCol := map[string]string{
		"uri":     "repo.uri",
		"name":    "repo.name",
		"created": "repo.created_at",
		"updated": "repo.updated_at",
		"pushed":  "repo.pushed_at",
	}
	if sortCol, valid := sortKeyToCol[sort]; valid {
		sort = sortCol
	} else {
		return "", nil, &sourcegraph.InvalidOptionsError{Reason: "invalid sort: " + sort}
	}

	direction := opt.Direction
	if direction == "" {
		direction = "asc"
	}
	if direction != "asc" && direction != "desc" {
		return "", nil, &sourcegraph.InvalidOptionsError{Reason: "invalid direction: " + direction}
	}
	orderBySQL += fmt.Sprintf("%s %s NULLS LAST", sort, direction)

	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE %s ORDER BY %s`, selectSQL, fromSQL, whereSQL, orderBySQL)

	return sql, args, nil
}

func (s *Repos) query(ctx context.Context, sql string, args ...interface{}) ([]*sourcegraph.Repo, error) {
	var repos []*dbRepo
	err := dbh(ctx).Select(&repos, sql, args...)
	if err != nil {
		return nil, err
	}
	return toRepos(repos), nil
}

func (s *Repos) Create(ctx context.Context, newRepo *sourcegraph.Repo) (*sourcegraph.Repo, error) {
	// Explicitly created repos in the DB are mirrors because they
	// don't have a corresponding VCS repository on the filesystem for
	// them.
	if !newRepo.Mirror {
		return nil, store.ErrRepoMirrorOnly
	}

	if newRepo.HTTPCloneURL == "" && newRepo.SSHCloneURL == "" {
		return nil, store.ErrRepoNeedsCloneURL
	}

	if newRepo.DefaultBranch == "" {
		// TODO(sqs): set this in a layer above, not here (e.g., in
		// the NewRepo protobuf type).
		newRepo.DefaultBranch = "master"
	}

	var r dbRepo
	r.fromRepo(newRepo)
	if err := dbh(ctx).Insert(&r); err != nil {
		return nil, err
	}
	return r.toRepo(), nil
}

func (s *Repos) Update(ctx context.Context, op *sourcegraph.ReposUpdateOp) error {
	if op.Description != "" {
		_, err := dbh(ctx).Exec(`UPDATE repo SET "description"=$1 WHERE uri=$2`, strings.TrimSpace(op.Description), op.Repo.URI)
		if err != nil {
			return err
		}
	}
	if op.Language != "" {
		_, err := dbh(ctx).Exec(`UPDATE repo SET "language"=$1 WHERE uri=$2`, strings.TrimSpace(op.Language), op.Repo.URI)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Repos) Delete(ctx context.Context, repo string) error {
	_, err := dbh(ctx).Exec(`DELETE FROM repo WHERE uri=$1;`, repo)
	return err
}
