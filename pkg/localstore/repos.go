package localstore

import (
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"context"

	"github.com/lib/pq"
	"gopkg.in/gorp.v1"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/accesscontrol"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
)

var autoRepoWhitelist []*regexp.Regexp

func init() {
	AppSchema.CreateSQL = append(AppSchema.CreateSQL,
		`CREATE TABLE repo (
			id SERIAL PRIMARY KEY,
			uri citext,
			owner citext,
			name citext,
			description text,
			vcs text NOT NULL,
			http_clone_url text,
			ssh_clone_url text,
			homepage_url text,
			default_branch text NOT NULL,
			language text,
			blocked boolean,
			deprecated boolean,
			fork boolean,
			mirror boolean,
			private boolean,
			created_at timestamp with time zone,
			updated_at timestamp with time zone,
			pushed_at timestamp with time zone,
			vcs_synced_at timestamp with time zone,
			indexed_revision text,
			freeze_indexed_revision boolean,
			origin_repo_id text,
			origin_service integer,
			origin_api_base_url text
		)`,
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
	AppSchema.DropSQL = append(AppSchema.DropSQL, "DROP TABLE repo")

	for _, pattern := range strings.Fields(env.Get("AUTO_REPO_WHITELIST", ".+", "whitelist of repositories that will be automatically added to the DB when opened (space-separated list of lower-case regular expressions)")) {
		expr, err := regexp.Compile("^" + pattern + "$")
		if err != nil {
			log.Fatalf("invalid regular expression %q in AUTO_REPO_WHITELIST: %s", pattern, err)
		}
		autoRepoWhitelist = append(autoRepoWhitelist, expr)
	}
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

	repos, err := s.getBySQL(ctx, "WHERE id=$1 LIMIT 1", id)
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, ErrRepoNotFound
	}
	repo := repos[0]

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
// If the repository doesn't already exist in the db, this method will
// add it to the db if the repo exists and start cloning, but will
// not wait for cloning to finish before returning.
//
// If the repository already exists in the db, that information is returned
// and no effort is made to detect if the repo is cloned or cloning.
func (s *repos) GetByURI(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	if Mocks.Repos.GetByURI != nil {
		return Mocks.Repos.GetByURI(ctx, uri)
	}

	repo, err := s.getByURI(ctx, uri)
	if err != nil {
		whitelisted := false
		for _, expr := range autoRepoWhitelist {
			if expr.MatchString(strings.ToLower(uri)) {
				whitelisted = true
				break
			}
		}
		if !whitelisted {
			return nil, err
		}

		if strings.HasPrefix(strings.ToLower(uri), "github.com/") {
			// Repo does not exist in DB, create new entry.
			ctx = context.WithValue(ctx, github.GitHubTrackingContextKey, "Repos.GetByURI")
			ghRepo, err := github.GetRepo(ctx, uri)
			if err != nil {
				return nil, err
			}
			if ghRepo.URI != uri {
				// not canonical name (the GitHub api will redirect from the old name to
				// the results for the new name if the repo got renamed on GitHub)
				if repo, err := s.getByURI(ctx, ghRepo.URI); err == nil {
					return repo, nil
				}
			}

			if err := s.TryInsertNew(ctx, ghRepo.URI, ghRepo.Description, ghRepo.Fork, ghRepo.Private); err != nil {
				return nil, err
			}

			return s.getByURI(ctx, ghRepo.URI)
		}

		return nil, err
	}

	return repo, nil
}

func (s *repos) getByURI(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	repos, err := s.getBySQL(ctx, "WHERE uri=$1 LIMIT 1", uri)
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, ErrRepoNotFound
	}
	repo := repos[0]

	// 🚨 SECURITY: access control check here 🚨
	if repo.Private && !verifyUserHasRepoURIAccess(ctx, repo.URI) {
		return nil, ErrRepoNotFound
	}

	return repo, nil
}

// getBySQL returns a repository matching the SQL query, if any
// exists. A "LIMIT 1" clause is appended to the query before it is
// executed.
func (s *repos) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.Repo, error) {
	rows, err := appDBH(ctx).Db.Query("SELECT id, uri, description, homepage_url, default_branch, language, blocked, fork, private, indexed_revision, created_at, updated_at, pushed_at, freeze_indexed_revision FROM repo "+query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []*sourcegraph.Repo
	for rows.Next() {
		var repo sourcegraph.Repo
		var freezeIndexedRevision *bool

		if err := rows.Scan(
			&repo.ID,
			&repo.URI,
			&repo.Description,
			&repo.HomepageURL,
			&repo.DefaultBranch,
			&repo.Language,
			&repo.Blocked,
			&repo.Fork,
			&repo.Private,
			&repo.IndexedRevision,
			&repo.CreatedAt,
			&repo.UpdatedAt,
			&repo.PushedAt,
			&freezeIndexedRevision,
		); err != nil {
			return nil, err
		}

		repo.FreezeIndexedRevision = freezeIndexedRevision != nil && *freezeIndexedRevision // FIXME: bad DB schema: nullable boolean

		repos = append(repos, &repo)
	}

	return repos, nil
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
	rawRepos, err := s.getBySQL(ctx, "WHERE "+strings.Join(conds, " AND ")+" LIMIT 1000", args...)

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

// UpdateRepoFieldsFromRemote sets the fields of the repository from the
// remote (e.g., GitHub) and updates the repository in the store layer.
func (s *repos) UpdateRepoFieldsFromRemote(ctx context.Context, repo *sourcegraph.Repo) error {
	if strings.HasPrefix(strings.ToLower(repo.URI), "github.com/") {
		return s.updateRepoFieldsFromGitHub(ctx, repo)
	}
	return nil
}

func (s *repos) updateRepoFieldsFromGitHub(ctx context.Context, repo *sourcegraph.Repo) error {
	// Fetch latest metadata from GitHub
	ghrepo, err := github.GetRepo(ctx, repo.URI)
	if err != nil {
		return err
	}

	updated := false
	updateOp := &RepoUpdate{
		ReposUpdateOp: &sourcegraph.ReposUpdateOp{
			Repo: repo.ID,
		},
	}

	if ghrepo.URI != repo.URI {
		repo.URI = ghrepo.URI
		updateOp.URI = ghrepo.URI
		updated = true
	}
	if ghrepo.Description != repo.Description {
		repo.Description = ghrepo.Description
		updateOp.Description = ghrepo.Description
		updated = true
	}
	if ghrepo.HomepageURL != repo.HomepageURL {
		repo.HomepageURL = ghrepo.HomepageURL
		updateOp.HomepageURL = ghrepo.HomepageURL
		updated = true
	}
	if ghrepo.DefaultBranch != repo.DefaultBranch {
		repo.DefaultBranch = ghrepo.DefaultBranch
		updateOp.DefaultBranch = ghrepo.DefaultBranch
		updated = true
	}
	if ghrepo.Language != repo.Language {
		repo.Language = ghrepo.Language
		updateOp.Language = ghrepo.Language
		updated = true
	}
	if ghrepo.Blocked != repo.Blocked {
		repo.Blocked = ghrepo.Blocked
		if ghrepo.Blocked {
			updateOp.Blocked = sourcegraph.ReposUpdateOp_TRUE
		} else {
			updateOp.Blocked = sourcegraph.ReposUpdateOp_FALSE
		}
		updated = true
	}
	if ghrepo.Fork != repo.Fork {
		repo.Fork = ghrepo.Fork
		if ghrepo.Fork {
			updateOp.Fork = sourcegraph.ReposUpdateOp_TRUE
		} else {
			updateOp.Fork = sourcegraph.ReposUpdateOp_FALSE
		}
		updated = true
	}
	uintPtrEqual := func(a, b *uint) bool {
		return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
	}
	if !uintPtrEqual(ghrepo.StarsCount, repo.StarsCount) {
		repo.StarsCount = ghrepo.StarsCount
	}
	if !uintPtrEqual(ghrepo.ForksCount, repo.ForksCount) {
		repo.ForksCount = ghrepo.ForksCount
	}
	if ghrepo.Private != repo.Private {
		repo.Private = ghrepo.Private
		if ghrepo.Private {
			updateOp.Private = sourcegraph.ReposUpdateOp_TRUE
		} else {
			updateOp.Private = sourcegraph.ReposUpdateOp_FALSE
		}
		updated = true
	}

	if !timestampEqual(repo.UpdatedAt, ghrepo.UpdatedAt) {
		repo.UpdatedAt = ghrepo.UpdatedAt
		updateOp.UpdatedAt = ghrepo.UpdatedAt
		updated = true
	}
	if !timestampEqual(repo.PushedAt, ghrepo.PushedAt) {
		repo.PushedAt = ghrepo.PushedAt
		updateOp.PushedAt = ghrepo.PushedAt
		updated = true
	}

	if !updated {
		return nil
	}

	log15.Debug("Updating repo metadata from remote", "repo", repo.URI)
	// UpdateRepoFieldsFromRemote is used in read requests, including
	// unauthed ones. However, this write isn't as the user, but
	// rather an optimization for us to save the data from
	// github. As such we use an elevated context to allow the
	// write.
	if err := s.Update(accesscontrol.WithInsecureSkip(ctx, true), *updateOp); err != nil {
		return err
	}

	return nil
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

// TryInsertNew attempts to insert the repository rp into the db. It returns no error if a repo
// with the given uri already exists.
func (s *repos) TryInsertNew(ctx context.Context, uri string, description string, fork bool, private bool) error {
	_, err := appDBH(ctx).Db.Exec("INSERT INTO repo (uri, description, fork, private, created_at, vcs, default_branch, homepage_url, language, blocked) VALUES ($1, $2, $3, $4, $5, '', '', '', '', false)", uri, description, fork, private, time.Now()) // FIXME: bad DB schema: nullable columns
	if err != nil {
		if isPQErrorUniqueViolation(err) {
			if c := err.(*pq.Error).Constraint; c == "repo_uri_unique" {
				return nil // repo with given uri already exists
			}
		}
		return err
	}
	return nil
}

func timestampEqual(a, b *time.Time) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}
