package db

import (
	"context"
	"errors"
	"fmt"
	regexpsyntax "regexp/syntax"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
)

type repoNotFoundErr struct {
	ID  api.RepoID
	URI api.RepoURI
}

func (e *repoNotFoundErr) Error() string {
	if e.URI != "" {
		return fmt.Sprintf("repo not found: uri=%q", e.URI)
	}
	if e.ID != 0 {
		return fmt.Sprintf("repo not found: id=%d", e.ID)
	}
	return "repo not found"
}

func (e *repoNotFoundErr) NotFound() bool {
	return true
}

// repos is a DB-backed implementation of the Repos
type repos struct{}

// Get returns metadata for the request repository ID. It fetches data
// only from the database and NOT from any external sources. If the
// caller is concerned the copy of the data in the database might be
// stale, the caller is responsible for fetching data from any
// external services.
func (s *repos) Get(ctx context.Context, id api.RepoID) (*types.Repo, error) {
	if Mocks.Repos.Get != nil {
		return Mocks.Repos.Get(ctx, id)
	}

	repos, err := s.getBySQL(ctx, sqlf.Sprintf("WHERE id=%d LIMIT 1", id))
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, &repoNotFoundErr{ID: id}
	}
	return repos[0], nil
}

// GetByURI returns the repository with the given URI from the database, or an
// error. If the repo doesn't exist in the DB, then errcode.IsNotFound will
// return true on the error returned. It does not attempt to look up or update
// the repository on any external service (such as its code host).
func (s *repos) GetByURI(ctx context.Context, uri api.RepoURI) (*types.Repo, error) {
	if Mocks.Repos.GetByURI != nil {
		return Mocks.Repos.GetByURI(ctx, uri)
	}

	repos, err := s.getBySQL(ctx, sqlf.Sprintf("WHERE uri=%s LIMIT 1", uri))
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, &repoNotFoundErr{URI: uri}
	}
	return repos[0], nil
}

func (s *repos) Count(ctx context.Context, opt ReposListOptions) (int, error) {
	if Mocks.Repos.Count != nil {
		return Mocks.Repos.Count(ctx, opt)
	}

	conds, err := s.listSQL(opt)
	if err != nil {
		return 0, err
	}

	q := sqlf.Sprintf("SELECT COUNT(*) FROM repo WHERE %s", sqlf.Join(conds, "AND"))

	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *repos) getBySQL(ctx context.Context, querySuffix *sqlf.Query) ([]*types.Repo, error) {
	q := sqlf.Sprintf("SELECT id, uri, description, language, enabled, indexed_revision, created_at, updated_at, freeze_indexed_revision, external_id, external_service_type, external_service_id FROM repo %s", querySuffix)
	rows, err := dbconn.Global.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []*types.Repo
	for rows.Next() {
		var repo types.Repo
		var freezeIndexedRevision *bool
		var spec dbExternalRepoSpec

		if err := rows.Scan(
			&repo.ID,
			&repo.URI,
			&repo.Description,
			&repo.Language,
			&repo.Enabled,
			&repo.IndexedRevision,
			&repo.CreatedAt,
			&repo.UpdatedAt,
			&freezeIndexedRevision,
			&spec.id, &spec.serviceType, &spec.serviceID,
		); err != nil {
			return nil, err
		}

		repo.FreezeIndexedRevision = freezeIndexedRevision != nil && *freezeIndexedRevision // FIXME: bad DB schema: nullable boolean
		repo.ExternalRepo = spec.toAPISpec()

		repos = append(repos, &repo)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return repos, nil
}

// ReposListOptions specifies the options for listing repositories.
//
// Query and IncludePatterns/ExcludePatterns may not be used together.
type ReposListOptions struct {
	// Query specifies a search query for repositories. If specified, then the Sort and
	// Direction options are ignored
	Query string

	// IncludePatterns is a list of regular expressions, all of which must match all
	// repositories returned in the list.
	IncludePatterns []string

	// ExcludePattern is a regular expression that must not match any repository
	// returned in the list.
	ExcludePattern string

	// Enabled includes enabled repositories in the list.
	Enabled bool

	// Disabled includes disabled repositories in the list.
	Disabled bool

	// NoForks excludes forks from the list.
	NoForks bool

	// OnlyForks excludes non-forks from the lhist.
	OnlyForks bool

	// NoArchived excludes archived repositories from the list.
	NoArchived bool

	// OnlyArchived excludes non-archived repositories from the list.
	OnlyArchived bool

	// Filters repositories based on whether they have an IndexedRevision set.
	HasIndexedRevision *bool

	// List of fields by which to order the return repositories.
	OrderBy RepoListOrderBy

	*LimitOffset
}

type RepoListOrderBy []RepoListSort

func (r RepoListOrderBy) SQL() *sqlf.Query {
	if len(r) == 0 {
		return sqlf.Sprintf(`ORDER BY id ASC`)
	}

	clauses := make([]*sqlf.Query, 0, len(r))
	for _, s := range r {
		clauses = append(clauses, s.SQL())
	}
	return sqlf.Sprintf(`ORDER BY %s`, sqlf.Join(clauses, ", "))
}

// RepoListSort is a field by which to sort and the direction of the sorting.
type RepoListSort struct {
	Field      RepoListColumn
	Descending bool
}

func (r RepoListSort) SQL() *sqlf.Query {
	if r.Descending {
		return sqlf.Sprintf(string(r.Field) + ` DESC`)
	}
	return sqlf.Sprintf(string(r.Field))
}

// RepoListColumn is a column by which repositories can be sorted. These correspond to columns in the database.
type RepoListColumn string

const (
	RepoListCreatedAt RepoListColumn = "created_at"
	RepoListURI       RepoListColumn = "uri"
)

// List lists repositories in the Sourcegraph repository
//
// This will not return any repositories from external services that are not present in the Sourcegraph repository.
// The result list is unsorted and has a fixed maximum limit of 1000 items.
// Matching is done with fuzzy matching, i.e. "query" will match any repo URI that matches the regexp `q.*u.*e.*r.*y`
func (s *repos) List(ctx context.Context, opt ReposListOptions) (results []*types.Repo, err error) {
	tr, ctx := trace.New(ctx, "repos.List", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if Mocks.Repos.List != nil {
		return Mocks.Repos.List(ctx, opt)
	}

	conds, err := s.listSQL(opt)
	if err != nil {
		return nil, err
	}

	// fetch matching repos
	fetchSQL := sqlf.Sprintf("WHERE %s %s %s", sqlf.Join(conds, "AND"), opt.OrderBy.SQL(), opt.LimitOffset.SQL())
	tr.LazyPrintf("SQL query: %s, SQL args: %v", fetchSQL.Query(sqlf.PostgresBindVar), fetchSQL.Args())
	rawRepos, err := s.getBySQL(ctx, fetchSQL)
	if err != nil {
		return nil, err
	}
	return rawRepos, nil
}

// ListEnabledNames returns a list of all enabled repo names. This is commonly
// requested information by other services (repo-updater and
// indexed-search). We special case just returning enabled names so that we
// read much less data into memory.
func (s *repos) ListEnabledNames(ctx context.Context) ([]string, error) {
	q := sqlf.Sprintf("SELECT uri FROM repo WHERE enabled = true")
	rows, err := dbconn.Global.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return names, nil
}

func (*repos) listSQL(opt ReposListOptions) (conds []*sqlf.Query, err error) {
	conds = []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opt.Query != "" && (len(opt.IncludePatterns) > 0 || opt.ExcludePattern != "") {
		return nil, errors.New("Repos.List: Query and IncludePatterns/ExcludePattern options are mutually exclusive")
	}
	if opt.Query != "" {
		conds = append(conds, sqlf.Sprintf("lower(uri) LIKE %s", "%"+strings.ToLower(opt.Query)+"%"))
	}
	for _, includePattern := range opt.IncludePatterns {
		exact, like, pattern, err := parseIncludePattern(includePattern)
		if err != nil {
			return nil, err
		}
		if exact != nil {
			if len(exact) == 0 || (len(exact) == 1 && exact[0] == "") {
				conds = append(conds, sqlf.Sprintf("TRUE"))
			} else {
				items := []*sqlf.Query{}
				for _, v := range exact {
					items = append(items, sqlf.Sprintf("%s", v))
				}
				conds = append(conds, sqlf.Sprintf("uri IN (%s)", sqlf.Join(items, ",")))
			}
		}
		if len(like) > 0 {
			var likeConds []*sqlf.Query
			for _, v := range like {
				likeConds = append(likeConds, sqlf.Sprintf(`lower(uri) LIKE %s`, strings.ToLower(v)))
			}
			conds = append(conds, sqlf.Sprintf("(%s)", sqlf.Join(likeConds, " OR ")))
		}
		if pattern != "" {
			conds = append(conds, sqlf.Sprintf("lower(uri) ~* %s", pattern))
		}
	}
	if opt.ExcludePattern != "" {
		conds = append(conds, sqlf.Sprintf("lower(uri) !~* %s", opt.ExcludePattern))
	}

	if opt.Enabled && opt.Disabled {
		// nothing to do
	} else if opt.Enabled && !opt.Disabled {
		conds = append(conds, sqlf.Sprintf("enabled"))
	} else if !opt.Enabled && opt.Disabled {
		conds = append(conds, sqlf.Sprintf("NOT enabled"))
	} else {
		return nil, errors.New("Repos.List: must specify at least one of Enabled=true or Disabled=true")
	}
	if opt.HasIndexedRevision != nil {
		if *opt.HasIndexedRevision {
			conds = append(conds, sqlf.Sprintf("indexed_revision IS NOT NULL"))
		} else {
			conds = append(conds, sqlf.Sprintf("indexed_revision IS NULL"))
		}
	}
	if opt.NoForks {
		conds = append(conds, sqlf.Sprintf("NOT fork"))
	}
	if opt.OnlyForks {
		conds = append(conds, sqlf.Sprintf("fork"))
	}
	if opt.NoArchived {
		conds = append(conds, sqlf.Sprintf("NOT archived"))
	}
	if opt.OnlyArchived {
		conds = append(conds, sqlf.Sprintf("archived"))
	}

	return conds, nil
}

// parseIncludePattern either (1) parses the pattern into a list of exact possible
// string values and LIKE patterns if such a list can be determined from the pattern,
// and (2) returns the original regexp if those patterns are not equivalent to the
// regexp.
//
// It allows Repos.List to optimize for the common case where a pattern like
// `(^github.com/foo/bar$)|(^github.com/baz/qux$)` is provided. In that case,
// it's faster to query for "WHERE uri IN (...)" the two possible exact values
// (because it can use an index) instead of using a "WHERE uri ~*" regexp condition
// (which generally can't use an index).
//
// This optimization is necessary for good performance when there are many repos
// in the database. With this optimization, specifying a "repogroup:" in the query
// will be fast (even if there are many repos) because the query can be constrained
// efficiently to only the repos in the group.
func parseIncludePattern(pattern string) (exact, like []string, regexp string, err error) {
	re, err := regexpsyntax.Parse(pattern, regexpsyntax.OneLine)
	if err != nil {
		return nil, nil, "", err
	}
	exact, contains, prefix, suffix, err := allMatchingStrings(re.Simplify())
	if err != nil {
		return nil, nil, "", err
	}
	for _, v := range contains {
		like = append(like, "%"+v+"%")
	}
	for _, v := range prefix {
		like = append(like, v+"%")
	}
	for _, v := range suffix {
		like = append(like, "%"+v)
	}
	if exact != nil || like != nil {
		return exact, like, "", nil
	}
	return nil, nil, pattern, nil
}

// allMatchingStrings returns a complete list of the strings that re
// matches, if it's possible to determine the list.
func allMatchingStrings(re *regexpsyntax.Regexp) (exact, contains, prefix, suffix []string, err error) {
	prog, err := regexpsyntax.Compile(re)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	switch {
	case re.Op == regexpsyntax.OpEmptyMatch:
		return []string{""}, nil, nil, nil, nil
	case re.Op == regexpsyntax.OpLiteral:
		prefix, complete := prog.Prefix()
		if complete {
			return nil, []string{prefix}, nil, nil, nil
		}
		return nil, nil, nil, nil, nil

	case re.Op == regexpsyntax.OpCharClass:
		// Only handle simple case of one range.
		if len(re.Rune) == 2 {
			len := int(re.Rune[1] - re.Rune[0] + 1)
			if len > 26 {
				// Avoid large character ranges (which could blow up the number
				// of possible matches).
				return nil, nil, nil, nil, nil
			}
			chars := make([]string, len)
			for r := re.Rune[0]; r <= re.Rune[1]; r++ {
				chars[r-re.Rune[0]] = string(r)
			}
			return nil, chars, nil, nil, nil
		}
		return nil, nil, nil, nil, nil

	case re.Op == regexpsyntax.OpBeginText:
		return nil, nil, []string{""}, nil, nil

	case re.Op == regexpsyntax.OpEndText:
		return nil, nil, nil, []string{""}, nil

	case re.Op == regexpsyntax.OpCapture:
		return allMatchingStrings(re.Sub0[0])

	case re.Op == regexpsyntax.OpConcat:
		var begin, end bool
		for i, sub := range re.Sub {
			if sub.Op == regexpsyntax.OpBeginText && i == 0 {
				begin = true
				continue
			}
			if sub.Op == regexpsyntax.OpEndText && i == len(re.Sub)-1 {
				end = true
				continue
			}
			subexact, subcontains, subprefix, subsuffix, err := allMatchingStrings(sub)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			if subexact == nil && subcontains == nil && subprefix == nil && subsuffix == nil {
				return nil, nil, nil, nil, nil
			}

			if subexact == nil {
				subexact = subcontains
			}
			if exact == nil {
				exact = subexact
			} else {
				size := len(exact) * len(subexact)
				if len(subexact) > 4 || size > 30 {
					// Avoid blowup in number of possible matches.
					return nil, nil, nil, nil, nil
				}
				combined := make([]string, 0, size)
				for _, match := range exact {
					for _, submatch := range subexact {
						combined = append(combined, match+submatch)
					}
				}
				exact = combined
			}
		}
		if exact == nil {
			exact = []string{""}
		}
		if begin && end {
			return exact, nil, nil, nil, nil
		} else if begin {
			return nil, nil, exact, nil, nil
		} else if end {
			return nil, nil, nil, exact, nil
		}
		return nil, exact, nil, nil, nil

	case re.Op == regexpsyntax.OpAlternate:
		for _, sub := range re.Sub {
			subexact, subcontains, subprefix, subsuffix, err := allMatchingStrings(sub)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			exact = append(exact, subexact...)
			contains = append(contains, subcontains...)
			prefix = append(prefix, subprefix...)
			suffix = append(suffix, subsuffix...)
		}
		return exact, contains, prefix, suffix, nil
	}

	return nil, nil, nil, nil, nil
}

// Delete deletes the repository row from the repo table. It will also delete any rows in the GlobalDeps and Pkgs stores
// that reference the deleted repository row.
func (s *repos) Delete(ctx context.Context, repo api.RepoID) error {
	if Mocks.Repos.Delete != nil {
		return Mocks.Repos.Delete(ctx, repo)
	}

	// Delete entries in pkgs and global_dep tables that correspond to the repo first
	if err := GlobalDeps.Delete(ctx, repo); err != nil {
		return err
	}
	if err := Pkgs.Delete(ctx, repo); err != nil {
		return err
	}

	// Hard delete entries in the discussions tables that correspond to the repo.
	threads, err := DiscussionThreads.List(ctx, &DiscussionThreadsListOptions{
		TargetRepoID: &repo,
	})
	if err != nil {
		return err
	}
	for _, thread := range threads {
		_, err := DiscussionThreads.Update(ctx, thread.ID, &DiscussionThreadsUpdateOptions{hardDelete: true})
		if err != nil {
			return err
		}
	}

	q := sqlf.Sprintf("DELETE FROM REPO WHERE id=%d", repo)
	_, err = dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	return err
}

func (s *repos) SetEnabled(ctx context.Context, id api.RepoID, enabled bool) error {
	q := sqlf.Sprintf("UPDATE repo SET enabled=%t WHERE id=%d", enabled, id)
	res, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return &repoNotFoundErr{ID: id}
	}
	return nil
}

func (s *repos) UpdateLanguage(ctx context.Context, repo api.RepoID, language string) error {
	_, err := dbconn.Global.ExecContext(ctx, "UPDATE repo SET language=$1 WHERE id=$2", language, repo)
	return err
}

func (s *repos) UpdateIndexedRevision(ctx context.Context, repo api.RepoID, commitID api.CommitID) error {
	_, err := dbconn.Global.ExecContext(ctx, "UPDATE repo SET indexed_revision=$1 WHERE id=$2", commitID, repo)
	return err
}

func (s *repos) UpdateRepositoryMetadata(ctx context.Context, uri api.RepoURI, description string, fork bool, archived bool) error {
	_, err := dbconn.Global.ExecContext(ctx, "UPDATE repo SET description=$1, fork=$2, archived=$3 WHERE uri=$4 	AND (description <> $1 OR fork <> $2 OR archived <> $3)", description, fork, archived, uri)
	return err
}

const upsertSQL = `WITH UPSERT AS (
	UPDATE repo SET uri=$1, description=$2, fork=$3, enabled=$4, external_id=$5, external_service_type=$6, external_service_id=$7, archived=$9 WHERE uri=$1 RETURNING uri
)
INSERT INTO repo(uri, description, fork, language, enabled, external_id, external_service_type, external_service_id, archived) (
	SELECT $1 AS uri, $2 AS description, $3 AS fork, $8 as language, $4 AS enabled,
	       $5 AS external_id, $6 AS external_service_type, $7 AS external_service_id, $9 AS archived
	WHERE $1 NOT IN (SELECT uri FROM upsert)
)`

// Upsert updates the repository if it already exists (keyed on URI) and
// inserts it if it does not.
//
// If repo exists, op.Enabled is ignored.
func (s *repos) Upsert(ctx context.Context, op api.InsertRepoOp) error {
	if Mocks.Repos.Upsert != nil {
		return Mocks.Repos.Upsert(op)
	}

	insert := false
	language := ""
	enabled := op.Enabled

	// We optimistically assume the repo is already in the table, so first
	// check if it is. We then fallback to the upsert functionality. The
	// upsert is logged as a modification to the DB, even if it is a no-op. So
	// we do this check to avoid log spam if postgres is configured with
	// log_statement='mod'.
	r, err := s.GetByURI(ctx, op.URI)
	if err != nil {
		if _, ok := err.(*repoNotFoundErr); !ok {
			return err
		}
		insert = true // missing
	} else {
		enabled = r.Enabled
		language = r.Language
		// Ignore Enabled for deciding to update
		insert = ((op.Description != r.Description) ||
			(op.Fork != r.Fork) ||
			(!op.ExternalRepo.Equal(r.ExternalRepo)))
	}

	if !insert {
		return nil
	}

	spec := (&dbExternalRepoSpec{}).fromAPISpec(op.ExternalRepo)
	_, err = dbconn.Global.ExecContext(ctx, upsertSQL, op.URI, op.Description, op.Fork, enabled, spec.id, spec.serviceType, spec.serviceID, language, op.Archived)
	return err
}

// dbExternalRepoSpec is convenience type for inserting or selecting *api.ExternalRepoSpec database data.
type dbExternalRepoSpec struct{ id, serviceType, serviceID *string }

func (s *dbExternalRepoSpec) fromAPISpec(spec *api.ExternalRepoSpec) *dbExternalRepoSpec {
	if spec != nil {
		*s = dbExternalRepoSpec{
			id:          &spec.ID,
			serviceType: &spec.ServiceType,
			serviceID:   &spec.ServiceID,
		}
	}
	return s
}

func (s dbExternalRepoSpec) toAPISpec() *api.ExternalRepoSpec {
	if s.id != nil && s.serviceType != nil && s.serviceID != nil {
		return &api.ExternalRepoSpec{
			ID:          *s.id,
			ServiceType: *s.serviceType,
			ServiceID:   *s.serviceID,
		}
	}
	return nil
}
