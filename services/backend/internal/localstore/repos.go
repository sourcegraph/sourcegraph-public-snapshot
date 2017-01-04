package localstore

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"context"

	"github.com/lib/pq"
	"gopkg.in/gorp.v1"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
)

// TODO remove skipFS by decoupling packages
var skipFS = false // used by tests

func init() {
	AppSchema.Map.AddTableWithName(dbRepo{}, "repo").SetKeys(true, "ID")
	AppSchema.CreateSQL = append(AppSchema.CreateSQL,
		"ALTER TABLE repo ALTER COLUMN uri TYPE citext",
		"ALTER TABLE repo ALTER COLUMN owner TYPE citext", // migration 2016.9.30
		"ALTER TABLE repo ALTER COLUMN name TYPE citext",  // migration 2016.9.30
		"CREATE UNIQUE INDEX repo_uri_unique ON repo(uri);",
		"ALTER TABLE repo ALTER COLUMN description TYPE text",
		`ALTER TABLE repo ALTER COLUMN default_branch SET NOT NULL;`,
		`ALTER TABLE repo ALTER COLUMN vcs SET NOT NULL;`,
		`ALTER TABLE repo ALTER COLUMN updated_at TYPE timestamp with time zone USING updated_at::timestamp with time zone;`,
		`ALTER TABLE repo ALTER COLUMN pushed_at TYPE timestamp with time zone USING pushed_at::timestamp with time zone;`,
		`ALTER TABLE repo ALTER COLUMN vcs_synced_at TYPE timestamp with time zone USING vcs_synced_at::timestamp with time zone;`,
		"CREATE INDEX repo_name ON repo(name text_pattern_ops);",

		"CREATE INDEX repo_owner_ci ON repo(owner);", // migration 2016.9.30
		"CREATE INDEX repo_name_ci ON repo(name);",   // migration 2016.9.30

		// migration 2016.9.30: `DROP INDEX repo_lower_uri_lower_name;`
	)
}

// dbRepo DB-maps a sourcegraph.Repo object.
type dbRepo struct {
	ID              int32
	URI             string
	Owner           string
	Name            string
	Description     string
	VCS             string
	HTTPCloneURL    string `db:"http_clone_url"`
	SSHCloneURL     string `db:"ssh_clone_url"`
	HomepageURL     string `db:"homepage_url"`
	DefaultBranch   string `db:"default_branch"`
	Language        string
	Blocked         bool
	Deprecated      bool
	Fork            bool
	Mirror          bool
	Private         bool
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       *time.Time `db:"updated_at"`
	PushedAt        *time.Time `db:"pushed_at"`
	VCSSyncedAt     *time.Time `db:"vcs_synced_at"`
	IndexedRevision *string    `db:"indexed_revision"`

	OriginRepoID     *string `db:"origin_repo_id"`
	OriginService    *int32  `db:"origin_service"` // values from Origin.ServiceType enum
	OriginAPIBaseURL *string `db:"origin_api_base_url"`
}

func (r *dbRepo) toRepo() *sourcegraph.Repo {
	r2 := &sourcegraph.Repo{
		ID:              r.ID,
		URI:             r.URI,
		Owner:           r.Owner,
		Name:            r.Name,
		Description:     r.Description,
		HTTPCloneURL:    r.HTTPCloneURL,
		SSHCloneURL:     r.SSHCloneURL,
		HomepageURL:     r.HomepageURL,
		DefaultBranch:   r.DefaultBranch,
		Language:        r.Language,
		Blocked:         r.Blocked,
		Deprecated:      r.Deprecated,
		Fork:            r.Fork,
		Private:         r.Private,
		IndexedRevision: r.IndexedRevision,
	}

	r2.CreatedAt = &r.CreatedAt
	r2.UpdatedAt = r.UpdatedAt
	r2.PushedAt = r.PushedAt
	r2.VCSSyncedAt = r.VCSSyncedAt
	return r2
}

func (r *dbRepo) fromRepo(r2 *sourcegraph.Repo) {
	r.ID = r2.ID
	r.URI = r2.URI
	r.Owner = r2.Owner
	r.Name = r2.Name
	r.Description = r2.Description
	r.HTTPCloneURL = r2.HTTPCloneURL
	r.SSHCloneURL = r2.SSHCloneURL
	r.HomepageURL = r2.HomepageURL
	r.DefaultBranch = r2.DefaultBranch
	r.Language = r2.Language
	r.Blocked = r2.Blocked
	r.Deprecated = r2.Deprecated
	r.Fork = r2.Fork
	r.Private = r2.Private
	if r2.CreatedAt != nil {
		r.CreatedAt = *r2.CreatedAt
	}
	r.UpdatedAt = r2.UpdatedAt
	r.PushedAt = r2.PushedAt
	r.VCSSyncedAt = r2.VCSSyncedAt
	r.IndexedRevision = r2.IndexedRevision
}

func toRepos(rs []*dbRepo) []*sourcegraph.Repo {
	r2s := make([]*sourcegraph.Repo, len(rs))
	for i, r := range rs {
		r2s[i] = r.toRepo()
	}
	return r2s
}

// repos is a DB-backed implementation of the Repos
type repos struct{}

// Get returns metadata for the request repository ID. It fetches data
// only from the database and NOT from any external sources. If the
// caller is concerned the copy of the data in the database might be
// stale, the caller is responsible for fetching data from any
// external services.
func (s *repos) Get(ctx context.Context, id int32) (*sourcegraph.Repo, error) {
	if Mocks.Repos.Get != nil {
		return Mocks.Repos.Get(ctx, id)
	}

	repo, err := s.getBySQL(ctx, "id=$1", id)
	if err != nil {
		return nil, err
	}
	// Check permissions against GitHub. The reason this does not call `VerifyUserHasReadAccess` is because
	// VerifyUserHasReadAccess --calls-> getRepo --calls-> repos.Get
	// TODO: Try to get rid of that loop, and have a way to use VerifyUserHasReadAccess here.
	if !accesscontrol.VerifyActorHasRepoURIAccess(ctx, auth.ActorFromContext(ctx), "Repos.Get", repo.ID, repo.URI) {
		return nil, accesscontrol.ErrRepoNotFound
	}
	return repo, nil
}

// GetByURI returns metadata for the request repository URI. See the
// documentation for repos.Get for the contract on the freshness of
// the data returned.
func (s *repos) GetByURI(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	if Mocks.Repos.GetByURI != nil {
		return Mocks.Repos.GetByURI(ctx, uri)
	}

	repo, err := s.getByURI(ctx, uri)
	if err != nil {
		return nil, err
	}
	// Check permissions against GitHub. The reason this does not call `VerifyUserHasReadAccess` is because
	// VerifyUserHasReadAccess --calls-> getRepo --calls-> repos.GetByURI
	// TODO: Try to get rid of that loop, and have a way to use VerifyUserHasReadAccess here.
	if !accesscontrol.VerifyActorHasRepoURIAccess(ctx, auth.ActorFromContext(ctx), "Repos.GetByURI", repo.ID, repo.URI) {
		return nil, accesscontrol.ErrRepoNotFound
	}
	return repo, nil
}

// SECURITY: DO NOT USE THIS METHOD! It does not verify access to the repo with
// GitHub and thus is very dangerous. It should only be called by permission
// checking functions in the accesscontrol package.
func (s *repos) UnsafeDangerousGetByURI(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	return s.getByURI(ctx, uri)
}

func (s *repos) getByURI(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	repo, err := s.getBySQL(ctx, "uri=$1", uri)
	if err != nil {
		if legacyerr.ErrCode(err) == legacyerr.NotFound {
			// Overwrite with error message containing repo URI.
			err = legacyerr.Errorf(legacyerr.NotFound, "%s: %s", err, uri)
		}
	}
	return repo, err
}

// getBySQL returns a repository matching the SQL query, if any
// exists. A "LIMIT 1" clause is appended to the query before it is
// executed.
func (s *repos) getBySQL(ctx context.Context, query string, args ...interface{}) (*sourcegraph.Repo, error) {
	var repo dbRepo
	if err := appDBH(ctx).SelectOne(&repo, "SELECT * FROM repo WHERE ("+query+") LIMIT 1", args...); err == sql.ErrNoRows {
		return nil, accesscontrol.ErrRepoNotFound
	} else if err != nil {
		return nil, err
	}
	return repo.toRepo(), nil
}

type RepoListOp struct {
	// Name filters the repository list by name.
	Name string

	// Query specifies a search query for repositories. If specified, then the Sort and
	// Direction options are ignored
	Query string

	// URIs specifies a set of repository URIs to list.
	URIs []string

	// Sort determines how the returned list of repositories is sorted.
	Sort string

	// Direction determines the sort direction.
	Direction string

	// NoFork excludes forks from the list of returned repositories.
	NoFork bool

	// Type of repositories to list. Possible values are currently
	// ones supported by the GitHub API, including: all, owner,
	// public, private, member. Default is "all".
	Type string

	// Owner filters the list of repositories to those with the
	// specified owner.
	Owner string

	sourcegraph.ListOptions
}

// List repositories in the Sourcegraph repository  Note:
// this will not return any repositories from external services
// that are not present in the Sourcegraph repository
func (s *repos) List(ctx context.Context, opt *RepoListOp) ([]*sourcegraph.Repo, error) {
	if Mocks.Repos.List != nil {
		return Mocks.Repos.List(ctx, opt)
	}

	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Repos.List", nil); err != nil {
		return nil, err
	}
	if opt == nil {
		opt = &RepoListOp{}
	}

	sql, args, err := reposListSQL(opt)
	if err != nil {
		return nil, err
	}
	rawRepos, err := s.query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	// SECURITY: It is very important that the input list of repos (rawRepos)
	// comes directly from the DB as VerifyUserHasReadAccessAll relies directly
	// on the accuracy of the Repo.Private field.
	repos, err := accesscontrol.VerifyUserHasReadAccessAll(ctx, "Repos.List", rawRepos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

type priorityRepo struct {
	priority int
	*sourcegraph.Repo
}

type priorityRepoList struct {
	repos []*priorityRepo
}

func (repos *priorityRepoList) Len() int {
	return len(repos.repos)
}

func (repos *priorityRepoList) Swap(i, j int) {
	repos.repos[i], repos.repos[j] = repos.repos[j], repos.repos[i]
}

func (repos *priorityRepoList) Less(i, j int) bool {
	return repos.repos[i].priority > repos.repos[j].priority
}

var errOptionsSpecifyEmptyResult = errors.New("pgsql: options specify and empty result set")

var (
	repoQuerySplitter    = regexp.MustCompile(`[/\s]+`)
	repoOwnerRepoPattern = regexp.MustCompile(`^([^/\s]+)[/\s]+([^\s]+)$`)
	repoOwnerPattern     = regexp.MustCompile(`^([^/\s]+)[/\s]+$`)
)

// reposListSQL translates the options struct to the SQL for querying
// PosgreSQL.
func reposListSQL(opt *RepoListOp) (string, []interface{}, error) {
	var selectSQL, fromSQL, whereSQL, orderBySQL string

	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	var queryTerms []string
	for _, v := range repoQuerySplitter.Split(opt.Query, -1) {
		if v != "" {
			queryTerms = append(queryTerms, v)
		}
	}

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
		if strings.Contains(opt.Query, "/") && len(queryTerms) >= 1 {
			fields := queryTerms
			if queryTerms[0] == "github.com" && len(fields) > 1 {
				fields = queryTerms[1:]
			}
			conds = append(conds, `owner=`+arg(fields[0]))
			if len(fields) > 1 {
				conds = append(conds, "name ILIKE "+arg(fields[1]+"%"))
			}
		} else if len(queryTerms) >= 1 {
			var queryConds []string
			for _, queryTerm := range queryTerms {
				queryConds = append(queryConds, `name=`+arg(queryTerm), `owner=`+arg(queryTerm))
			}
			conds = append(conds, fmt.Sprintf(`(%s)`, strings.Join(queryConds, " OR ")))
		}
		switch opt.Type {
		case "private":
			conds = append(conds, `private`)
		case "public":
			conds = append(conds, `NOT private`)
		case "", "all":
		default:
			return "", nil, legacyerr.Errorf(legacyerr.InvalidArgument, "invalid state")
		}
		if opt.Owner != "" {
			conds = append(conds, `lower(owner)=`+arg(strings.ToLower(opt.Owner)))
		}

		if conds != nil {
			whereSQL = "(" + strings.Join(conds, ") AND (") + ")"
		} else {
			whereSQL = "true"
		}
	}

	// ORDER BY
	var orderByTerms []string
	if match := repoOwnerPattern.FindAllStringSubmatch(opt.Query, 1); len(match) == 1 && len(match[0]) == 2 {
		// "$OWNER/" case
		orderByTerms = append(orderByTerms, `owner=`+arg(match[0][1])+` DESC`)
	} else if match := repoOwnerRepoPattern.FindAllStringSubmatch(opt.Query, 1); len(match) == 1 && len(match[0]) == 3 {
		// "$OWNER/$REPO" case
		orderByTerms = append(orderByTerms, `(owner=`+arg(match[0][1])+` AND `+`name=`+arg(match[0][2])+`) DESC`)
	}
	orderByTerms = append(orderByTerms, "NOT fork DESC")
	if len(queryTerms) >= 1 {
		// rank repositories with a name equaling the last search term higher
		orderByTerms = append(orderByTerms, "name="+arg(queryTerms[len(queryTerms)-1])+" DESC")
	}
	if len(queryTerms) >= 2 {
		orderByTerms = append(orderByTerms, "owner="+arg(queryTerms[len(queryTerms)-2])+" DESC")
	}
	if len(queryTerms) >= 2 {
		// Prefix matching for repo name.
		last := queryTerms[len(queryTerms)-1]
		orderByTerms = append(orderByTerms, "name ILIKE "+arg(last+"%")+" DESC")
	}
	orderByTerms = append(orderByTerms, "private DESC", "name ASC")

	sort := opt.Sort
	if sort == "" {
		sort = "id"
	}

	sortKeyToCol := map[string]string{
		"id":      "repo.id",
		"uri":     "repo.uri",
		"path":    "repo.uri",
		"name":    "repo.name",
		"created": "repo.created_at",
		"updated": "repo.updated_at",
		"pushed":  "repo.pushed_at",
	}
	if sortCol, valid := sortKeyToCol[sort]; valid {
		sort = sortCol
	} else {
		return "", nil, legacyerr.Errorf(legacyerr.InvalidArgument, "invalid sort: "+sort)
	}

	direction := opt.Direction
	if direction == "" {
		direction = "asc"
	}
	if direction != "asc" && direction != "desc" {
		return "", nil, legacyerr.Errorf(legacyerr.InvalidArgument, "invalid direction: "+direction)
	}
	orderByTerms = append(orderByTerms, fmt.Sprintf("%s %s NULLS LAST", sort, direction))
	orderBySQL = strings.Join(orderByTerms, ", ")

	// LIMIT
	limitSQL := arg(opt.Limit())
	offsetSQL := arg(opt.Offset())

	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE %s ORDER BY %s LIMIT %s OFFSET %s`, selectSQL, fromSQL, whereSQL, orderBySQL, limitSQL, offsetSQL)
	return sql, args, nil
}

func (s *repos) query(ctx context.Context, sql string, args ...interface{}) ([]*sourcegraph.Repo, error) {
	var repos []*dbRepo
	if _, err := appDBH(ctx).Select(&repos, sql, args...); err != nil {
		return nil, err
	}
	return toRepos(repos), nil
}

// Create a repository and return its ID.
func (s *repos) Create(ctx context.Context, newRepo *sourcegraph.Repo) (int32, error) {
	if Mocks.Repos.Create != nil {
		return Mocks.Repos.Create(ctx, newRepo)
	}

	if strings.HasPrefix(newRepo.URI, "github.com/") {
		// Anyone can create GitHub mirrors.
	} else if strings.HasPrefix(newRepo.URI, "source.developers.google.com/p/") {
		// Anyone can create GCP mirrors.
	} else if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Repos.Create", nil); err != nil {
		return 0, err
	}

	if repo, err := s.getByURI(ctx, newRepo.URI); err == nil {
		return 0, legacyerr.Errorf(legacyerr.AlreadyExists, "repo already exists: %s", repo.URI)
	}

	var r dbRepo
	r.fromRepo(newRepo)
	err := appDBH(ctx).Insert(&r)
	if isPQErrorUniqueViolation(err) {
		if c := err.(*pq.Error).Constraint; c != "repo_uri_unique" {
			log15.Warn("Expected unique_violation of repo_uri_unique constraint, but it was something else; did it change?", "constraint", c, "err", err)
		}
		return 0, legacyerr.Errorf(legacyerr.AlreadyExists, "repo already exists: %s", newRepo.URI)
	}
	return r.ID, err
}

// RepoUpdate represents an update to specific fields of a repo. Only
// fields with non-zero values are updated.
//
// The ReposUpdateOp.Repo field must be filled in to specify the repo
// that will be updated.
type RepoUpdate struct {
	*sourcegraph.ReposUpdateOp

	UpdatedAt *time.Time
	PushedAt  *time.Time
}

// Update a repository.
func (s *repos) Update(ctx context.Context, op RepoUpdate) error {
	if Mocks.Repos.Update != nil {
		return Mocks.Repos.Update(ctx, op)
	}

	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Repos.Update", op.Repo); err != nil {
		return err
	}

	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	var updates []string
	if op.URI != "" {
		updates = append(updates, `"uri"=`+arg(op.URI))
	}
	if op.Owner != "" {
		updates = append(updates, `"owner"=`+arg(op.Owner))
	}
	if op.Name != "" {
		updates = append(updates, `"name"=`+arg(op.Name))
	}
	if op.Description != "" {
		updates = append(updates, `"description"=`+arg(op.Description))
	}
	if op.HTTPCloneURL != "" {
		updates = append(updates, `"http_clone_url"=`+arg(op.HTTPCloneURL))
	}
	if op.SSHCloneURL != "" {
		updates = append(updates, `"ssh_clone_url"=`+arg(op.SSHCloneURL))
	}
	if op.HomepageURL != "" {
		updates = append(updates, `"homepage_url"=`+arg(op.HomepageURL))
	}
	if op.DefaultBranch != "" {
		updates = append(updates, `"default_branch"=`+arg(op.DefaultBranch))
	}
	if op.Language != "" {
		updates = append(updates, `"language"=`+arg(op.Language))
	}
	if op.Blocked != sourcegraph.ReposUpdateOp_NONE {
		updates = append(updates, `"blocked"=`+arg(op.Blocked == sourcegraph.ReposUpdateOp_TRUE))
	}
	if op.Deprecated != sourcegraph.ReposUpdateOp_NONE {
		updates = append(updates, `"deprecated"=`+arg(op.Deprecated == sourcegraph.ReposUpdateOp_TRUE))
	}
	if op.Fork != sourcegraph.ReposUpdateOp_NONE {
		updates = append(updates, `"fork"=`+arg(op.Fork == sourcegraph.ReposUpdateOp_TRUE))
	}
	if op.Mirror != sourcegraph.ReposUpdateOp_NONE {
		updates = append(updates, `"mirror"=`+arg(op.Mirror == sourcegraph.ReposUpdateOp_TRUE))
	}
	if op.Private != sourcegraph.ReposUpdateOp_NONE {
		updates = append(updates, `"private"=`+arg(op.Private == sourcegraph.ReposUpdateOp_TRUE))
	}
	if op.UpdatedAt != nil {
		updates = append(updates, `"updated_at"=`+arg(op.UpdatedAt))
	}
	if op.PushedAt != nil {
		updates = append(updates, `"pushed_at"=`+arg(op.PushedAt))
	}
	if op.IndexedRevision != "" {
		updates = append(updates, `"indexed_revision"=`+arg(op.IndexedRevision))
	}

	if len(updates) > 0 {
		sql := `UPDATE repo SET ` + strings.Join(updates, ", ") + ` WHERE id=` + arg(op.Repo)
		_, err := appDBH(ctx).Exec(sql, args...)
		return err
	}
	return nil
}

// InternalRepoUpdate is an update of repo fields that are used by
// internal Sourcegraph processes only. It is separate from RepoUpdate
// so that internal updates can be performed by machine processes that
// do not need to assume the privileges of the repo's owner (and can
// merely have an internal token scope).
type InternalRepoUpdate struct {
	VCSSyncedAt *time.Time
}

// InternalUpdate performs an update of internal repository
// fields. See InternalRepoUpdate for more information.
func (s *repos) InternalUpdate(ctx context.Context, repo int32, op InternalRepoUpdate) error {
	if Mocks.Repos.InternalUpdate != nil {
		return Mocks.Repos.InternalUpdate(ctx, repo, op)
	}

	// SECURITY NOTE: If you add more fields and more UPDATE queries,
	// each one should perform its own access checks, since updating
	// different fields may require different levels of
	// privilege. Here, we check for read access, which is the minimum
	// privilege level that any InternalUpdate call must require.
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Repos.InternalRepoUpdate", repo); err != nil {
		return err
	}

	if op.VCSSyncedAt != nil {
		// SECURITY NOTE: Even though this operation causes a DB
		// write, we only require read access, since we are merely
		// updating the date when the VCS data was synced.
		if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Repos.InternalRepoUpdate", repo); err != nil {
			return err
		}

		_, err := appDBH(ctx).Exec(`UPDATE repo SET "vcs_synced_at"=$1 WHERE id=$2`, op.VCSSyncedAt, repo)
		if err != nil {
			return err
		}
	}
	return nil
}
