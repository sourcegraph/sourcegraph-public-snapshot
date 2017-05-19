package localstore

import (
	"database/sql"
	"sort"
	"strings"
	"time"

	"context"

	"github.com/lib/pq"
	"gopkg.in/gorp.v1"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/accesscontrol"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/gitcmd"
)

func init() {
	AppSchema.Map.AddTableWithName(dbRepoOrig{}, "repo").SetKeys(true, "ID")
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
		"CREATE INDEX repo_uri_trgm ON repo USING GIN (lower(uri) gin_trgm_ops);",

		// migration 2016.9.30: `DROP INDEX repo_lower_uri_lower_name;`
	)
}

// dbRepo DB-maps a sourcegraph.Repo object. The reason for the split
// between dbRepo and dbRepoOrig is to support adding new columns to
// the table without downtime or a heavyweight migration.
//
// Fields in dbRepo (i.e., NOT in dbRepoOrig) should be nullable (in
// both the data structure and the DB) and are not set on new repo
// creation.
//
// Adding new fields to the repo table should follow these steps:
// - Add the field to dbRepo. Test that it works both with and without
//   the new field in the database.
// - Deploy the new code into production.
// - Add the new column to the repo table in production.
// - Move the field from dbRepo into dbRepoOrig and post announcement
//   to other devs of the migration (they now need to run it in dev).
type dbRepo struct {
	dbRepoOrig
	Canonical *bool `db:"canonical"`
}

type dbRepoOrig struct {
	ID                    int32
	URI                   string
	Owner                 string
	Name                  string
	Description           string
	VCS                   string
	HTTPCloneURL          string `db:"http_clone_url"`
	SSHCloneURL           string `db:"ssh_clone_url"`
	HomepageURL           string `db:"homepage_url"`
	DefaultBranch         string `db:"default_branch"`
	Language              string
	Blocked               bool
	Deprecated            bool
	Fork                  bool
	Mirror                bool
	Private               bool
	CreatedAt             time.Time  `db:"created_at"`
	UpdatedAt             *time.Time `db:"updated_at"`
	PushedAt              *time.Time `db:"pushed_at"`
	VCSSyncedAt           *time.Time `db:"vcs_synced_at"`
	IndexedRevision       *string    `db:"indexed_revision"`
	FreezeIndexedRevision *bool      `db:"freeze_indexed_revision"`

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
		HomepageURL:     r.HomepageURL,
		DefaultBranch:   r.DefaultBranch,
		Language:        r.Language,
		Blocked:         r.Blocked,
		Fork:            r.Fork,
		Private:         r.Private,
		IndexedRevision: r.IndexedRevision,
	}

	if r.FreezeIndexedRevision != nil && *r.FreezeIndexedRevision {
		r2.FreezeIndexedRevision = true
	}
	if r.Canonical != nil && *r.Canonical {
		r2.Canonical = true
	}

	r2.CreatedAt = &r.CreatedAt
	r2.UpdatedAt = r.UpdatedAt
	r2.PushedAt = r.PushedAt
	return r2
}

func (r *dbRepo) fromRepo(r2 *sourcegraph.Repo) {
	r.ID = r2.ID
	r.URI = r2.URI
	r.Owner = r2.Owner
	r.Name = r2.Name
	r.Description = r2.Description
	r.HomepageURL = r2.HomepageURL
	r.DefaultBranch = r2.DefaultBranch
	r.Language = r2.Language
	r.Blocked = r2.Blocked
	r.Fork = r2.Fork
	r.Private = r2.Private
	if r2.CreatedAt != nil {
		r.CreatedAt = *r2.CreatedAt
	}
	r.UpdatedAt = r2.UpdatedAt
	r.PushedAt = r2.PushedAt
	r.IndexedRevision = r2.IndexedRevision

	var freezeIndexedRevision bool = r2.FreezeIndexedRevision
	r.FreezeIndexedRevision = &freezeIndexedRevision

	var canonical = r2.Canonical
	r.Canonical = &canonical
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
	// 🚨 SECURITY: access control check here 🚨
	if repo.Private && !verifyUserHasRepoURIAccess(ctx, repo.URI) {
		return nil, ErrRepoNotFound
	}
	return repo, nil
}

// GetByURI returns metadata for the request repository URI. See the
// documentation for repos.Get for the contract on the freshness of
// the data returned.
//
// If the requested repository does not exist, this method will
// attempt to clone it and add it to the DB. In the interim, it will
// return an error that says something like "Not found (clone in
// progress)". Callers can poll until the returned error is nil. If
// the new candidate repository could not be cloned, no repository is
// added to the DB.
//
// Note that if this method returns a repository and non-nil error, it
// does not necessarily mean the repository has been cloned, as the
// repository in the DB could have been added by other means (e.g.,
// manually).
func (s *repos) GetByURI(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	if Mocks.Repos.GetByURI != nil {
		return Mocks.Repos.GetByURI(ctx, uri)
	}

	repo, err := s.getByURI(ctx, uri)
	if err != nil {
		var newRepo *sourcegraph.Repo
		if github.IsRepoAndShouldCheckPermissions(uri) {
			// Repo does not exist in DB, create new entry.
			ctx = context.WithValue(ctx, github.GitHubTrackingContextKey, "Repos.GetByURI")
			ghRepo, err := github.GetRepo(ctx, uri)
			if err != nil {
				return nil, err
			}
			ghRepoURI := githubutil.RepoURI(ghRepo.Owner, ghRepo.Name)
			if ghRepoURI != uri {
				// not canonical name (the GitHub api will redirect from the old name to
				// the results for the new name if the repo got renamed on GitHub)
				if repo, err := s.getByURI(ctx, uri); err == nil {
					return repo, nil
				}
			}

			// Purposefully set very few fields. We don't want to cache
			// metadata, because it'll get stale, and fetching online from
			// GitHub is quite easy and (with HTTP caching) performant.
			ts := time.Now()
			newRepo = &sourcegraph.Repo{
				Owner:       ghRepo.Owner,
				Name:        ghRepo.Name,
				URI:         ghRepoURI,
				Description: ghRepo.Description,
				Fork:        ghRepo.Fork,
				Private:     ghRepo.Private,
				CreatedAt:   &ts,
			}
		} else {
			ts := time.Now()
			cmps := strings.Split(uri, "/")
			var owner string
			if len(cmps) >= 2 {
				owner = cmps[len(cmps)-2]
			}
			newRepo = &sourcegraph.Repo{
				Name:      cmps[len(cmps)-1],
				Owner:     owner,
				URI:       uri,
				CreatedAt: &ts,
			}
		}
		// If gitserver has not yet cloned it, don't add it to the DB
		if _, err := gitcmd.Open(newRepo).ResolveRevision(ctx, "HEAD"); err != nil {
			return nil, err
		}
		return s.tryInsertNew(ctx, newRepo)
	}
	return repo, nil
}

func (s *repos) getByURI(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	repo, err := s.getBySQL(ctx, "uri=$1", uri)
	if err != nil {
		if legacyerr.ErrCode(err) == legacyerr.NotFound {
			// Overwrite with error message containing repo URI.
			err = legacyerr.Errorf(legacyerr.NotFound, "%s: %s", err, uri)
		}
		return nil, err
	}

	// 🚨 SECURITY: access control check here 🚨
	if repo.Private && !verifyUserHasRepoURIAccess(ctx, repo.URI) {
		return nil, ErrRepoNotFound
	}

	return repo, nil
}

// getBySQL returns a repository matching the SQL query, if any
// exists. A "LIMIT 1" clause is appended to the query before it is
// executed.
func (s *repos) getBySQL(ctx context.Context, query string, args ...interface{}) (*sourcegraph.Repo, error) {
	var repo dbRepo
	if err := appDBH(ctx).SelectOne(&repo, "SELECT * FROM repo WHERE ("+query+") LIMIT 1", args...); err == sql.ErrNoRows {
		return nil, ErrRepoNotFound
	} else if err != nil {
		return nil, err
	}
	return repo.toRepo(), nil
}

type RepoListOp struct {
	// Query specifies a search query for repositories. If specified, then the Sort and
	// Direction options are ignored
	Query string

	sourcegraph.ListOptions
}

// List repositories in the Sourcegraph repository  Note:
// this will not return any repositories from external services
// that are not present in the Sourcegraph repository
func (s *repos) List(ctx context.Context, opt *RepoListOp) ([]*sourcegraph.Repo, error) {
	if Mocks.Repos.List != nil {
		return Mocks.Repos.List(ctx, opt)
	}

	if opt == nil {
		opt = &RepoListOp{}
	}

	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	terms := strings.Fields(opt.Query)
	if len(terms) > 10 {
		terms = terms[:10]
	}

	conds := []string{"TRUE"}
	for _, term := range terms {
		term = strings.ToLower(term)
		term = strings.Replace(term, `\`, `\\`, -1)
		term = strings.Replace(term, "%", `\%`, -1)
		term = strings.Replace(term, "_", `\_`, -1)
		conds = append(conds, "lower(uri) LIKE "+arg("%"+term+"%"))
	}

	// fetch matching repos unordered
	sql := "SELECT * FROM repo WHERE " + strings.Join(conds, " AND ") + " LIMIT 1000;"
	rawRepos, err := s.query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: It is very important that the input list of repos (rawRepos) 🚨
	// comes directly from the DB as verifyUserHasReadAccessAll relies directly
	// on the accuracy of the Repo.Private field.
	repos, err := verifyUserHasReadAccessAll(ctx, "Repos.List", rawRepos)
	if err != nil {
		return nil, err
	}

	// sort by position of search terms
	sort.Slice(repos, func(i, j int) bool {
		uri1 := strings.ToLower(repos[i].URI)
		uri2 := strings.ToLower(repos[j].URI)
		for _, term := range terms {
			term = strings.ToLower(term)
			pos1 := strings.Index(uri1, term)
			pos2 := strings.Index(uri2, term)
			if pos1 < pos2 {
				return true
			}
			if pos2 < pos1 {
				return false
			}
		}
		return uri1 < uri2
	})

	// pagination
	if opt.Page > 0 {
		start := (opt.Page - 1) * opt.PerPage
		if int(start) >= len(repos) {
			return nil, nil
		}
		repos = repos[start:]
		if len(repos) > int(opt.PerPage) {
			repos = repos[:opt.PerPage]
		}
	}

	return repos, nil
}

func (s *repos) query(ctx context.Context, sql string, args ...interface{}) ([]*sourcegraph.Repo, error) {
	var repos []*dbRepo
	if _, err := appDBH(ctx).Select(&repos, sql, args...); err != nil {
		return nil, err
	}
	return toRepos(repos), nil
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

	if !accesscontrol.Skip(ctx) {
		return legacyerr.Errorf(legacyerr.PermissionDenied, "permission denied")
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
	if op.Fork != sourcegraph.ReposUpdateOp_NONE {
		updates = append(updates, `"fork"=`+arg(op.Fork == sourcegraph.ReposUpdateOp_TRUE))
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

// tryInsertNew attempts to insert the repository rp into the db. On
// success, it returns the repository that was actually inserted.
func (s *repos) tryInsertNew(ctx context.Context, rp *sourcegraph.Repo) (*sourcegraph.Repo, error) {
	var r dbRepo
	r.fromRepo(rp)
	if err := appDBH(ctx).Insert(&r.dbRepoOrig); err != nil {
		if isPQErrorUniqueViolation(err) {
			if c := err.(*pq.Error).Constraint; c != "repo_uri_unique" {
				log15.Warn("Expected unique_violation of repo_uri_unique constraint, but it was something else; did it change?", "constraint", c, "err", err)
			}
			return s.getByURI(ctx, rp.URI) // might be race condition, try to read
		}
		return nil, err
	}
	return r.toRepo(), nil
}
