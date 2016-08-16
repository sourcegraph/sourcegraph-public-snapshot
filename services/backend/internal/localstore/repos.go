package localstore

import (
	"database/sql"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"context"

	"github.com/lib/pq"
	"github.com/neelance/parallel"
	gogithub "github.com/sourcegraph/go-github/github"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/gorp.v1"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	approuter "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
	"sourcegraph.com/sqs/pbtypes"
)

// TODO remove skipFS by decoupling packages
var skipFS = false // used by tests

func init() {
	AppSchema.Map.AddTableWithName(dbRepo{}, "repo").SetKeys(true, "ID")
	AppSchema.CreateSQL = append(AppSchema.CreateSQL,
		"ALTER TABLE repo ALTER COLUMN uri TYPE citext",
		"CREATE UNIQUE INDEX repo_uri_unique ON repo(uri);",
		"ALTER TABLE repo ALTER COLUMN description TYPE text",
		`ALTER TABLE repo ALTER COLUMN default_branch SET NOT NULL;`,
		`ALTER TABLE repo ALTER COLUMN vcs SET NOT NULL;`,
		`ALTER TABLE repo ALTER COLUMN updated_at TYPE timestamp with time zone USING updated_at::timestamp with time zone;`,
		`ALTER TABLE repo ALTER COLUMN pushed_at TYPE timestamp with time zone USING pushed_at::timestamp with time zone;`,
		`ALTER TABLE repo ALTER COLUMN vcs_synced_at TYPE timestamp with time zone USING vcs_synced_at::timestamp with time zone;`,
		"CREATE INDEX repo_name ON repo(name text_pattern_ops);",

		// fast for repo searching by URI and name
		"CREATE INDEX repo_lower_uri_lower_name ON repo((lower(uri)::text) text_pattern_ops, lower(name));",
	)
}

// dbRepo DB-maps a sourcegraph.Repo object.
type dbRepo struct {
	ID            int32
	URI           string
	Owner         string
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
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at"`
	PushedAt      *time.Time `db:"pushed_at"`
	VCSSyncedAt   *time.Time `db:"vcs_synced_at"`

	OriginRepoID     *string `db:"origin_repo_id"`
	OriginService    *int32  `db:"origin_service"` // values from protobuf Origin.ServiceType enum
	OriginAPIBaseURL *string `db:"origin_api_base_url"`
}

func (r *dbRepo) toRepo() *sourcegraph.Repo {
	r2 := &sourcegraph.Repo{
		ID:            r.ID,
		URI:           r.URI,
		Owner:         r.Owner,
		Name:          r.Name,
		Description:   r.Description,
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
	}

	{
		ts := pbtypes.NewTimestamp(r.CreatedAt)
		r2.CreatedAt = &ts
	}
	if r.UpdatedAt != nil {
		ts := pbtypes.NewTimestamp(*r.UpdatedAt)
		r2.UpdatedAt = &ts
	}
	if r.PushedAt != nil {
		ts := pbtypes.NewTimestamp(*r.PushedAt)
		r2.PushedAt = &ts
	}
	if r.VCSSyncedAt != nil {
		ts := pbtypes.NewTimestamp(*r.VCSSyncedAt)
		r2.VCSSyncedAt = &ts
	}
	if r.OriginRepoID != nil && r.OriginService != nil && r.OriginAPIBaseURL != nil {
		r2.Origin = &sourcegraph.Origin{
			ID:         *r.OriginRepoID,
			Service:    sourcegraph.Origin_ServiceType(*r.OriginService),
			APIBaseURL: *r.OriginAPIBaseURL,
		}
	}
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
	r.Mirror = r2.Mirror
	r.Private = r2.Private

	if r2.CreatedAt != nil {
		r.CreatedAt = r2.CreatedAt.Time()
	}
	if r2.UpdatedAt != nil {
		ts := r2.UpdatedAt.Time()
		r.UpdatedAt = &ts
	}
	if r2.PushedAt != nil {
		ts := r2.PushedAt.Time()
		r.PushedAt = &ts
	}
	if r2.VCSSyncedAt != nil {
		ts := r2.VCSSyncedAt.Time()
		r.VCSSyncedAt = &ts
	}
	if o := r2.Origin; o != nil {
		r.OriginRepoID = gogithub.String(o.ID)
		service := int32(o.Service)
		r.OriginService = &service
		r.OriginAPIBaseURL = gogithub.String(o.APIBaseURL)
	}
}

func toRepos(rs []*dbRepo) []*sourcegraph.Repo {
	r2s := make([]*sourcegraph.Repo, len(rs))
	for i, r := range rs {
		r2s[i] = r.toRepo()
	}
	return r2s
}

// repos is a DB-backed implementation of the Repos store.
type repos struct{}

var _ store.Repos = (*repos)(nil)

func (s *repos) Get(ctx context.Context, id int32) (*sourcegraph.Repo, error) {
	repo, err := s.getBySQL(ctx, "id=$1", id)
	if err != nil {
		return nil, err
	}
	return verifyAccessAndSetAllFields(ctx, repo)
}

func (s *repos) GetByURI(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	repo, err := s.getByURI(ctx, uri)
	if err != nil {
		return nil, err
	}
	return verifyAccessAndSetAllFields(ctx, repo)
}

// verifyAccessAndSetAllFields checks permissions and fills in additional fields on
// repo. It MUST be called after fetching a repo from the DB before
// returning the repo.
//
// NOTE: The repo returned is the same as the repo passed in. Provided as a
// convenience.
func verifyAccessAndSetAllFields(ctx context.Context, repo *sourcegraph.Repo) (*sourcegraph.Repo, error) {
	// Avoid an infinite loop (since
	// accesscontrol.VerifyUserHasReadAccess calls (*repos).Get).
	if strings.HasPrefix(strings.ToLower(repo.URI), "github.com/") {
		if err := accesscontrol.VerifyActorHasGitHubRepoAccess(ctx, auth.ActorFromContext(ctx), "Repos.Get", repo.ID, repo.URI); err != nil {
			return nil, err
		}

		// Automatically migrate repos that don't have their origin fields set.
		//
		// NOTE: This can be removed when all repos in the DB have their
		// origin fields set.
		if repo.Origin == nil && repo.ID != 0 {
			ghRepo, err := github.ReposFromContext(ctx).Get(ctx, repo.URI)
			if err != nil {
				// Highly unlikely to fail since the
				// VerifyActorHasGitHubRepoAccess would make this same
				// call, so let's treat it as fatal to catch any
				// possible rare edge cases.
				return nil, grpc.Errorf(codes.Internal, "while auto-migrating repo to add origin fields: %s", err)
			}
			if ghRepo.Origin != nil {
				if _, err := appDBH(ctx).Exec(`UPDATE repo SET origin_repo_id=$1, origin_service=$2, origin_api_base_url=$3 WHERE id=$4;`, ghRepo.Origin.ID, ghRepo.Origin.Service, "https://api.github.com", repo.ID); err != nil {
					return nil, err
				}
			}
		}
	} else {
		// All hosted repos or alternate URI repos are publicly viewable.
	}

	// TODO(keegancsmith) remove once we are storing all github metadata
	// in table https://app.asana.com/0/37478073567611/138332225969208
	if repo.DefaultBranch == "" {
		log15.Debug("Repo missing DefaultBranch", "repo", repo.URI)
		repo.DefaultBranch = "master"
	}

	// The clone field is not set in the DB since it would become stale if
	// the AppURL configuration changed.
	if !repo.Mirror {
		repo.HTTPCloneURL = conf.AppURL(ctx).ResolveReference(approuter.Rel.URLToRepo(repo.URI)).String()
	}

	return repo, nil
}

func (s *repos) getByURI(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	repo, err := s.getBySQL(ctx, "uri=$1", uri)
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			// Overwrite with error message containing repo URI.
			err = grpc.Errorf(codes.NotFound, "%s: %s", err, uri)
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
		return nil, grpc.Errorf(codes.NotFound, "repo not found") // can't nicely serialize args
	} else if err != nil {
		return nil, err
	}
	return repo.toRepo(), nil
}

func (s *repos) List(ctx context.Context, opt *sourcegraph.RepoListOptions) ([]*sourcegraph.Repo, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Repos.List", nil); err != nil {
		return nil, err
	}
	if opt == nil {
		opt = &sourcegraph.RepoListOptions{}
	}

	// Fetch GitHub repos that the user can access. We include these
	// repos in the result, and we pass these IDs to listFromDB
	// because it otherwise has no way of knowing (at the PostgreSQL
	// level) what repos the user can see.
	var accessibleGitHubRepoIDs []string
	ghRepos, err := s.listAllAccessibleGitHubRepos(ctx, opt)
	if err != nil {
		return nil, err
	}
	accessibleGitHubRepoIDs = make([]string, len(ghRepos))
	for i, r := range ghRepos {
		accessibleGitHubRepoIDs[i] = r.Origin.ID
	}

	// Fetch repos from the DB that are accessible to the user.
	// This will include all local repos that are not from "github.com".
	dbRepos, err := s.listFromDB(ctx, opt, accessibleGitHubRepoIDs)
	if err != nil {
		return nil, err
	}

	// Do access checks in parallel since it may do remote calls
	//
	// No need to do access checks for ghRepos, since we JUST fetched
	// them from GitHub.
	par := parallel.NewRun(30)
	for _, repo := range dbRepos {
		repo := repo
		par.Acquire()
		go func() {
			defer par.Release()
			if _, err := verifyAccessAndSetAllFields(ctx, repo); err != nil {
				par.Error(err)
			}
		}()
	}
	if err := par.Wait(); err != nil {
		return nil, err
	}

	dbReposOnGH := make(map[string]*sourcegraph.Repo, len(dbRepos)) // GitHub ID -> *sourcegraph.Repo in dbRepos.
	for _, dbRepo := range dbRepos {
		if o := dbRepo.Origin; o != nil && o.Service == sourcegraph.Origin_GitHub {
			dbReposOnGH[o.ID] = dbRepo
		}
	}

	var repos []*sourcegraph.Repo
	if !opt.RemoteOnly {
		// Combine results from GitHub (repos accessible to the current
		// user) and the DB (extra metadata, such as the repo's
		// Sourcegraph ID).

		// Include all accessible repos that are in DB.
		repos = dbRepos

		// Add GitHub repos that aren't already present in DB.
		for _, ghRepo := range ghRepos {
			if _, ok := dbReposOnGH[ghRepo.Origin.ID]; !ok {
				repos = append(repos, ghRepo)
			}
		}
	} else {
		// If opt.RemoteOnly is set, only return repos that are "accessible"
		// according to the remote, but augment them with extra metadata from DB.
		for _, ghRepo := range ghRepos {
			if db, ok := dbReposOnGH[ghRepo.Origin.ID]; ok {
				// Populate the Sourcegraph repo ID (from the database).
				ghRepo.ID = db.ID
			}
			repos = append(repos, ghRepo)
		}
	}

	return repos, nil
}

func (s *repos) listAllAccessibleGitHubRepos(ctx context.Context, opt *sourcegraph.RepoListOptions) ([]*sourcegraph.Repo, error) {
	// Only users have "accessible" repos.
	if !github.HasAuthedUser(ctx) {
		return nil, nil
	}

	var allRepos []*sourcegraph.Repo
	for page := 1; ; page++ {
		const perPage = 100
		repos, err := github.ReposFromContext(ctx).ListAccessible(ctx, &gogithub.RepositoryListOptions{
			Type: opt.Type,
			ListOptions: gogithub.ListOptions{
				PerPage: perPage,
				Page:    page,
			},
		})
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if len(repos) < perPage {
			break
		}
	}
	return allRepos, nil
}

func (s *repos) listFromDB(ctx context.Context, opt *sourcegraph.RepoListOptions, allowGitHubRepoIDs []string) ([]*sourcegraph.Repo, error) {
	sql, args, err := s.listSQL(opt, allowGitHubRepoIDs)
	if err != nil {
		if err == errOptionsSpecifyEmptyResult {
			err = nil
		}
		return nil, err
	}
	return s.query(ctx, sql, args...)
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

func (s *repos) Search(ctx context.Context, query string) ([]*sourcegraph.RepoSearchResult, error) {
	query = strings.TrimSpace(query)

	// Does not perform search with one character because the range is too broad.
	if len(query) < 2 {
		return []*sourcegraph.RepoSearchResult{}, nil
	}

	var exactArgs, fuzzArgs []interface{}
	exactArg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(exactArgs))
		exactArgs = append(exactArgs, a)
		return v
	}
	fuzzArg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(fuzzArgs))
		fuzzArgs = append(fuzzArgs, a)
		return v
	}

	// Perform exact match search first when possible,
	// do fuzz match search next if no exact match results found.
	performExactMatch := true
	baseSQL := "SELECT repo.* FROM repo WHERE fork=false AND"
	exactSQL, fuzzSQL := baseSQL, baseSQL

	// Values used for determine results' priority.
	var owner, name string

	// Slashes indicate the user knows exactly what they're looking for.
	if strings.Contains(query, "/") {
		fuzzSQL += fmt.Sprintf(" uri LIKE %s", fuzzArg("%"+query+"%"))

		fields := strings.Split(query, "/")
		if len(fields) == 2 {
			exactSQL += fmt.Sprintf(" LOWER(owner) = LOWER(%s) AND LOWER(name) = LOWER(%s)", exactArg(fields[0]), exactArg(fields[1]))
			owner, name = fields[0], fields[1]
		} else {
			performExactMatch = false
		}

	} else {
		fields := strings.Fields(query)
		if len(fields) == 1 {
			// Only one keyword, which could either be the owner or repo name.
			exactSQL += fmt.Sprintf(" (LOWER(owner) = LOWER(%s) OR LOWER(name) = LOWER(%s))", exactArg(query), exactArg(query))
			fuzzSQL += fmt.Sprintf(" (owner ILIKE %s OR name ILIKE %s)", fuzzArg(query+"%"), fuzzArg(query+"%"))
			owner, name = query, query

		} else if len(fields) == 2 {
			// Two keywords. The first could be owners, the second could be repo name.
			exactSQL += fmt.Sprintf(" LOWER(owner) = LOWER(%s) AND LOWER(name) = LOWER(%s)", exactArg(fields[0]), exactArg(fields[1]))
			fuzzSQL += fmt.Sprintf(" owner ILIKE %s AND name ILIKE %s", fuzzArg(fields[0]+"%"), fuzzArg(fields[1]+"%"))
			owner, name = fields[0], fields[1]

		} else {
			// Three keywords are too much.
			return []*sourcegraph.RepoSearchResult{}, nil
		}
	}

	exactSQL += " LIMIT 3"
	fuzzSQL += " LIMIT 3"

	var exactRepos, repos []*sourcegraph.Repo
	var err error
	if performExactMatch {
		exactRepos, err = s.query(ctx, exactSQL, exactArgs...)
		if err != nil {
			return nil, err
		}
	}

	if len(exactRepos) > 0 {
		repos = exactRepos
	} else {
		repos, err = s.query(ctx, fuzzSQL, fuzzArgs...)
		if err != nil {
			return nil, err
		}
	}

	priorityRepos := make([]*priorityRepo, 0, len(repos))
	for _, repo := range repos {
		prepo := &priorityRepo{
			Repo: repo,
		}
		if repo.Owner == owner && repo.Name == name {
			prepo.priority = 2
		} else if repo.Owner == owner || repo.Name == name {
			prepo.priority = 1
		}
		priorityRepos = append(priorityRepos, prepo)
	}

	sort.Sort(&priorityRepoList{repos: priorityRepos})

	// Critical permissions check. DO NOT REMOVE.
	var results []*sourcegraph.RepoSearchResult
	for _, prepo := range priorityRepos {
		if _, err := verifyAccessAndSetAllFields(ctx, prepo.Repo); err != nil {
			continue
		}
		results = append(results, &sourcegraph.RepoSearchResult{
			Repo: prepo.Repo,
		})
	}

	return results, nil
}

var errOptionsSpecifyEmptyResult = errors.New("pgsql: options specify and empty result set")

func (s *repos) listSQL(opt *sourcegraph.RepoListOptions, allowGitHubRepoIDs []string) (string, []interface{}, error) {
	var selectSQL, fromSQL, whereSQL, orderBySQL string

	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
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
			return "", nil, grpc.Errorf(codes.InvalidArgument, "invalid state")
		}
		if opt.Owner != "" {
			return "", nil, errOptionsSpecifyEmptyResult
		}

		{
			// Only allow List to return GitHub mirrors that have been
			// explicitly determined to be accessible. Our DB doesn't
			// cache the GitHub metadata, so we have no way of filtering
			// appropriately on any columns (including even just returning
			// public repos--what if they aren't public anymore?).
			cond := "uri NOT LIKE 'github.com/%'"
			if len(allowGitHubRepoIDs) > 0 {
				bvs := make([]string, len(allowGitHubRepoIDs))
				for i, v := range allowGitHubRepoIDs {
					bvs[i] = arg(v)
				}
				cond = "(" + cond + " OR (origin_service=" + arg(sourcegraph.Origin_GitHub) + " AND origin_repo_id IN (" + strings.Join(bvs, ",") + ")))"
			}
			conds = append(conds, cond)
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
		return "", nil, grpc.Errorf(codes.InvalidArgument, "invalid sort: "+sort)
	}

	direction := opt.Direction
	if direction == "" {
		direction = "asc"
	}
	if direction != "asc" && direction != "desc" {
		return "", nil, grpc.Errorf(codes.InvalidArgument, "invalid direction: "+direction)
	}
	orderBySQL += fmt.Sprintf("%s %s NULLS LAST", sort, direction)

	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE %s ORDER BY %s`, selectSQL, fromSQL, whereSQL, orderBySQL)
	return sql, args, nil
}

func (s *repos) query(ctx context.Context, sql string, args ...interface{}) ([]*sourcegraph.Repo, error) {
	var repos []*dbRepo
	if _, err := appDBH(ctx).Select(&repos, sql, args...); err != nil {
		return nil, err
	}
	return toRepos(repos), nil
}

func (s *repos) Create(ctx context.Context, newRepo *sourcegraph.Repo) (int32, error) {
	if strings.HasPrefix(newRepo.URI, "github.com/") {
		if !newRepo.Mirror {
			return 0, grpc.Errorf(codes.InvalidArgument, "cannot create hosted repo with URI prefix: 'github.com/'")
		}
		// Anyone can create GitHub mirrors.
	} else if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Repos.Create", nil); err != nil {
		return 0, err
	}

	if repo, err := s.getByURI(ctx, newRepo.URI); err == nil {
		return 0, grpc.Errorf(codes.AlreadyExists, "repo already exists: %s", repo.URI)
	}

	// Create the filesystem repo where the git data lives. (The repo
	// metadata, such as the existence, description, language, etc.,
	// live in PostgreSQL.)
	// A mirrored repo is automatically cloned by the repo updater instead of here.
	if !newRepo.Mirror && !skipFS {
		if err := gitserver.Init(newRepo.URI); err != nil && err != vcs.ErrRepoExist {
			return 0, err
		}
	}

	var r dbRepo
	r.fromRepo(newRepo)
	err := appDBH(ctx).Insert(&r)
	if isPQErrorUniqueViolation(err) {
		if c := err.(*pq.Error).Constraint; c != "repo_uri_unique" {
			log15.Warn("Expected unique_violation of repo_uri_unique constraint, but it was something else; did it change?", "constraint", c, "err", err)
		}
		return 0, grpc.Errorf(codes.AlreadyExists, "repo already exists: %s", newRepo.URI)
	}
	return r.ID, err
}

func (s *repos) Update(ctx context.Context, op store.RepoUpdate) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Repos.Update", op.Repo); err != nil {
		return err
	}
	if op.Description != "" {
		_, err := appDBH(ctx).Exec(`UPDATE repo SET "description"=$1 WHERE id=$2`, strings.TrimSpace(op.Description), op.Repo)
		if err != nil {
			return err
		}
	}
	if op.Language != "" {
		_, err := appDBH(ctx).Exec(`UPDATE repo SET "language"=$1 WHERE id=$2`, strings.TrimSpace(op.Language), op.Repo)
		if err != nil {
			return err
		}
	}
	if op.DefaultBranch != "" {
		_, err := appDBH(ctx).Exec(`UPDATE repo SET "default_branch"=$1 WHERE id=$2`, strings.TrimSpace(op.DefaultBranch), op.Repo)
		if err != nil {
			return err
		}
	}
	if op.Fork != sourcegraph.ReposUpdateOp_NONE {
		_, err := appDBH(ctx).Exec(`UPDATE repo SET "fork"=$1 WHERE id=$2`, op.Fork == sourcegraph.ReposUpdateOp_TRUE, op.Repo)
		if err != nil {
			return err
		}
	}
	if op.Private != sourcegraph.ReposUpdateOp_NONE {
		_, err := appDBH(ctx).Exec(`UPDATE repo SET "private"=$1 WHERE id=$2`, op.Private == sourcegraph.ReposUpdateOp_TRUE, op.Repo)
		if err != nil {
			return err
		}
	}

	if op.UpdatedAt != nil {
		_, err := appDBH(ctx).Exec(`UPDATE repo SET "updated_at"=$1 WHERE id=$2`, op.UpdatedAt, op.Repo)
		if err != nil {
			return err
		}
	}
	if op.PushedAt != nil {
		_, err := appDBH(ctx).Exec(`UPDATE repo SET "pushed_at"=$1 WHERE id=$2`, op.PushedAt, op.Repo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *repos) InternalUpdate(ctx context.Context, repo int32, op store.InternalRepoUpdate) error {
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

func (s *repos) Delete(ctx context.Context, repo int32) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Repos.Delete", repo); err != nil {
		return err
	}

	var dir string
	if !skipFS {
		var err error
		dir, err = getRepoDir(ctx, repo)
		if err != nil {
			return err
		}
	}

	_, err := appDBH(ctx).Exec(`DELETE FROM repo WHERE id=$1;`, repo)
	if err != nil {
		return err
	}
	if !skipFS && dir != "" {
		if err := gitserver.Remove(dir); err != nil {
			log15.Warn("Deleting repo on filesystem failed", "repo", repo, "dir", dir, "err", err)
		}
	}
	return nil
}
