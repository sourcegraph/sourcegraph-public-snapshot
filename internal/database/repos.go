package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	regexpsyntax "github.com/grafana/regexp/syntax"
	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/pagure"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/perforce"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/phabricator"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepoNotFoundErr struct {
	ID         api.RepoID
	Name       api.RepoName
	HashedName api.RepoHashedName
}

func (e *RepoNotFoundErr) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("repo not found: name=%q", e.Name)
	}
	if e.ID != 0 {
		return fmt.Sprintf("repo not found: id=%d", e.ID)
	}
	return "repo not found"
}

func (e *RepoNotFoundErr) NotFound() bool {
	return true
}

type RepoStore interface {
	basestore.ShareableStore[schemas.Production]
	Transact(context.Context) (RepoStore, error)
	With(basestore.ShareableStore[schemas.Production]) RepoStore
	Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error)
	Done(error) error

	Count(context.Context, ReposListOptions) (int, error)
	Create(context.Context, ...*types.Repo) error
	Delete(context.Context, ...api.RepoID) error
	Get(context.Context, api.RepoID) (*types.Repo, error)
	GetByIDs(context.Context, ...api.RepoID) ([]*types.Repo, error)
	GetByName(context.Context, api.RepoName) (*types.Repo, error)
	GetByHashedName(context.Context, api.RepoHashedName) (*types.Repo, error)
	GetFirstRepoNameByCloneURL(context.Context, string) (api.RepoName, error)
	GetReposSetByIDs(context.Context, ...api.RepoID) (map[api.RepoID]*types.Repo, error)
	List(context.Context, ReposListOptions) ([]*types.Repo, error)
	ListIndexableRepos(context.Context, ListIndexableReposOptions) ([]types.MinimalRepo, error)
	ListMinimalRepos(context.Context, ReposListOptions) ([]types.MinimalRepo, error)
	Metadata(context.Context, ...api.RepoID) ([]*types.SearchedRepo, error)
	StreamMinimalRepos(context.Context, ReposListOptions, func(*types.MinimalRepo)) error
}

var _ RepoStore = (*repoStore)(nil)

// repoStore handles access to the repo table
type repoStore struct {
	*basestore.Store[schemas.Production]
}

// ReposWith instantiates and returns a new RepoStore using the other
// store handle.
func ReposWith(other basestore.ShareableStore[schemas.Production]) RepoStore {
	return &repoStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *repoStore) With(other basestore.ShareableStore) RepoStore {
	return &repoStore{Store: s.Store.With(other)}
}

func (s *repoStore) Transact(ctx context.Context) (RepoStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &repoStore{Store: txBase}, err
}

// Get finds and returns the repo with the given repository ID from the database.
// When a repo isn't found or has been blocked, an error is returned.
func (s *repoStore) Get(ctx context.Context, id api.RepoID) (_ *types.Repo, err error) {
	tr, ctx := trace.New(ctx, "repos.Get", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	repos, err := s.listRepos(ctx, tr, ReposListOptions{
		IDs:            []api.RepoID{id},
		LimitOffset:    &LimitOffset{Limit: 1},
		IncludeBlocked: true,
	})
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, &RepoNotFoundErr{ID: id}
	}

	repo := repos[0]

	return repo, repo.IsBlocked()
}

var counterAccessGranted = promauto.NewCounter(prometheus.CounterOpts{
	Name: "src_access_granted_private_repo",
	Help: "metric to measure the impact of logging access granted to private repos",
})

func logPrivateRepoAccessGranted(ctx context.Context, db DB, ids []api.RepoID) {
	if disabled, _ := strconv.ParseBool(os.Getenv("SRC_DISABLE_LOG_PRIVATE_REPO_ACCESS")); disabled {
		return
	}

	a := actor.FromContext(ctx)
	arg, _ := json.Marshal(struct {
		Resource string       `json:"resource"`
		Service  string       `json:"service"`
		Repos    []api.RepoID `json:"repo_ids"`
	}{
		Resource: "db.repo",
		Service:  env.MyName,
		Repos:    ids,
	})

	event := &SecurityEvent{
		Name:            SecurityEventNameAccessGranted,
		URL:             "",
		UserID:          uint32(a.UID),
		AnonymousUserID: "",
		Argument:        arg,
		Source:          "BACKEND",
		Timestamp:       time.Now(),
	}

	// If this event was triggered by an internal actor we need to ensure that at
	// least the UserID or AnonymousUserID field are set so that we don't trigger
	// the security_event_logs_check_has_user constraint
	if a.Internal {
		event.AnonymousUserID = "internal"
	}

	db.SecurityEventLogs().LogEvent(ctx, event)
}

// GetByName returns the repository with the given nameOrUri from the
// database, or an error. If we have a match on name and uri, we prefer the
// match on name.
//
// Name is the name for this repository (e.g., "github.com/user/repo"). It is
// the same as URI, unless the user configures a non-default
// repositoryPathPattern.
//
// When a repo isn't found or has been blocked, an error is returned.
func (s *repoStore) GetByName(ctx context.Context, nameOrURI api.RepoName) (_ *types.Repo, err error) {
	tr, ctx := trace.New(ctx, "repos.GetByName", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	repos, err := s.listRepos(ctx, tr, ReposListOptions{
		Names:          []string{string(nameOrURI)},
		LimitOffset:    &LimitOffset{Limit: 1},
		IncludeBlocked: true,
	})
	if err != nil {
		return nil, err
	}

	if len(repos) == 1 {
		return repos[0], repos[0].IsBlocked()
	}

	// We don't fetch in the same SQL query since uri is not unique and could
	// conflict with a name. We prefer returning the matching name if it
	// exists.
	repos, err = s.listRepos(ctx, tr, ReposListOptions{
		URIs:           []string{string(nameOrURI)},
		LimitOffset:    &LimitOffset{Limit: 1},
		IncludeBlocked: true,
	})
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, &RepoNotFoundErr{Name: nameOrURI}
	}

	return repos[0], repos[0].IsBlocked()
}

// GetByHashedName returns the repository with the given hashedName from the database, or an error.
// RepoHashedName is the repository hashed name.
// When a repo isn't found or has been blocked, an error is returned.
func (s *repoStore) GetByHashedName(ctx context.Context, repoHashedName api.RepoHashedName) (_ *types.Repo, err error) {
	tr, ctx := trace.New(ctx, "repos.GetByHashedName", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	repos, err := s.listRepos(ctx, tr, ReposListOptions{
		HashedName:     string(repoHashedName),
		LimitOffset:    &LimitOffset{Limit: 1},
		IncludeBlocked: true,
	})
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, &RepoNotFoundErr{HashedName: repoHashedName}
	}

	return repos[0], repos[0].IsBlocked()
}

// GetByIDs returns a list of repositories by given IDs. The number of results list could be less
// than the candidate list due to no repository is associated with some IDs.
func (s *repoStore) GetByIDs(ctx context.Context, ids ...api.RepoID) (_ []*types.Repo, err error) {
	tr, ctx := trace.New(ctx, "repos.GetByIDs", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// listRepos will return a list of all repos if we pass in an empty ID list,
	// so it is better to just return here rather than leak repo info.
	if len(ids) == 0 {
		return []*types.Repo{}, nil
	}
	return s.listRepos(ctx, tr, ReposListOptions{IDs: ids})
}

// GetReposSetByIDs returns a map of repositories with the given IDs, indexed by their IDs. The number of results
// entries could be less than the candidate list due to no repository is associated with some IDs.
func (s *repoStore) GetReposSetByIDs(ctx context.Context, ids ...api.RepoID) (map[api.RepoID]*types.Repo, error) {
	repos, err := s.GetByIDs(ctx, ids...)
	if err != nil {
		return nil, err
	}

	repoMap := make(map[api.RepoID]*types.Repo, len(repos))
	for _, r := range repos {
		repoMap[r.ID] = r
	}

	return repoMap, nil
}

func (s *repoStore) Count(ctx context.Context, opt ReposListOptions) (ct int, err error) {
	tr, ctx := trace.New(ctx, "repos.Count", "")
	defer func() {
		if err != nil {
			tr.SetError(err)
		}
		tr.Finish()
	}()

	opt.Select = []string{"COUNT(*)"}
	opt.OrderBy = nil
	opt.LimitOffset = nil

	err = s.list(ctx, tr, opt, func(rows *sql.Rows) error {
		return rows.Scan(&ct)
	})

	return ct, err
}

// Metadata returns repo metadata used to decorate search results. The returned slice may be smaller than the
// number of IDs given if a repo with the given ID does not exist.
func (s *repoStore) Metadata(ctx context.Context, ids ...api.RepoID) (_ []*types.SearchedRepo, err error) {
	tr, ctx := trace.New(ctx, "repos.Metadata", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	opts := ReposListOptions{
		IDs: ids,
		// Return a limited subset of fields
		Select: []string{
			"repo.id",
			"repo.name",
			"repo.description",
			"repo.fork",
			"repo.archived",
			"repo.private",
			"repo.stars",
			"gr.last_fetched",
		},
		// Required so gr.last_fetched is select-able
		joinGitserverRepos: true,
	}

	res := make([]*types.SearchedRepo, 0, len(ids))
	scanMetadata := func(rows *sql.Rows) error {
		var r types.SearchedRepo
		if err := rows.Scan(
			&r.ID,
			&r.Name,
			&dbutil.NullString{S: &r.Description},
			&r.Fork,
			&r.Archived,
			&r.Private,
			&dbutil.NullInt{N: &r.Stars},
			&r.LastFetched,
		); err != nil {
			return err
		}

		res = append(res, &r)
		return nil
	}

	return res, errors.Wrap(s.list(ctx, tr, opts, scanMetadata), "fetch metadata")
}

const listReposQueryFmtstr = `
-- source: internal/database/repos.go:list
%%s -- Populates "queryPrefix", i.e. CTEs
SELECT %s
FROM repo
%%s
WHERE
	%%s   -- Populates "queryConds"
	AND
	(%%s) -- Populates "authzConds"
%%s       -- Populates "querySuffix"
`

const getSourcesByRepoQueryStr = `
(
	SELECT
		json_agg(
		json_build_object(
			'CloneURL', esr.clone_url,
			'ID', esr.external_service_id,
			'Kind', LOWER(svcs.kind)
		)
		)
	FROM external_service_repos AS esr
	JOIN external_services AS svcs ON esr.external_service_id = svcs.id
	WHERE
		esr.repo_id = repo.id
		AND
		svcs.deleted_at IS NULL
)
`

var minimalRepoColumns = []string{
	"repo.id",
	"repo.name",
	"repo.private",
	"repo.stars",
}

var repoColumns = []string{
	"repo.id",
	"repo.name",
	"repo.private",
	"repo.external_id",
	"repo.external_service_type",
	"repo.external_service_id",
	"repo.uri",
	"repo.description",
	"repo.fork",
	"repo.archived",
	"repo.stars",
	"repo.created_at",
	"repo.updated_at",
	"repo.deleted_at",
	"repo.metadata",
	"repo.blocked",
}

func scanRepo(rows *sql.Rows, r *types.Repo) (err error) {
	var sources dbutil.NullJSONRawMessage
	var metadata json.RawMessage
	var blocked dbutil.NullJSONRawMessage

	err = rows.Scan(
		&r.ID,
		&r.Name,
		&r.Private,
		&dbutil.NullString{S: &r.ExternalRepo.ID},
		&dbutil.NullString{S: &r.ExternalRepo.ServiceType},
		&dbutil.NullString{S: &r.ExternalRepo.ServiceID},
		&dbutil.NullString{S: &r.URI},
		&dbutil.NullString{S: &r.Description},
		&r.Fork,
		&r.Archived,
		&dbutil.NullInt{N: &r.Stars},
		&r.CreatedAt,
		&dbutil.NullTime{Time: &r.UpdatedAt},
		&dbutil.NullTime{Time: &r.DeletedAt},
		&metadata,
		&blocked,
		&sources,
	)
	if err != nil {
		return err
	}

	if blocked.Raw != nil {
		r.Blocked = &types.RepoBlock{}
		if err = json.Unmarshal(blocked.Raw, r.Blocked); err != nil {
			return err
		}
	}

	type sourceInfo struct {
		ID       int64
		CloneURL string
		Kind     string
	}
	r.Sources = make(map[string]*types.SourceInfo)

	if sources.Raw != nil {
		var srcs []sourceInfo
		if err = json.Unmarshal(sources.Raw, &srcs); err != nil {
			return errors.Wrap(err, "scanRepo: failed to unmarshal sources")
		}
		for _, src := range srcs {
			urn := extsvc.URN(src.Kind, src.ID)
			r.Sources[urn] = &types.SourceInfo{
				ID:       urn,
				CloneURL: src.CloneURL,
			}
		}
	}

	typ, ok := extsvc.ParseServiceType(r.ExternalRepo.ServiceType)
	if !ok {
		log15.Warn("scanRepo - failed to parse service type", "r.ExternalRepo.ServiceType", r.ExternalRepo.ServiceType)
		return nil
	}
	switch typ {
	case extsvc.TypeGitHub:
		r.Metadata = new(github.Repository)
	case extsvc.TypeGitLab:
		r.Metadata = new(gitlab.Project)
	case extsvc.TypeGerrit:
		r.Metadata = new(gerrit.Project)
	case extsvc.TypeBitbucketServer:
		r.Metadata = new(bitbucketserver.Repo)
	case extsvc.TypeBitbucketCloud:
		r.Metadata = new(bitbucketcloud.Repo)
	case extsvc.TypeAWSCodeCommit:
		r.Metadata = new(awscodecommit.Repository)
	case extsvc.TypeGitolite:
		r.Metadata = new(gitolite.Repo)
	case extsvc.TypePerforce:
		r.Metadata = new(perforce.Depot)
	case extsvc.TypePhabricator:
		r.Metadata = new(phabricator.Repo)
	case extsvc.TypePagure:
		r.Metadata = new(pagure.Project)
	case extsvc.TypeOther:
		r.Metadata = new(extsvc.OtherRepoMetadata)
	case extsvc.TypeJVMPackages:
		r.Metadata = new(reposource.MavenMetadata)
	case extsvc.TypeNpmPackages:
		r.Metadata = new(reposource.NpmMetadata)
	case extsvc.TypeGoModules:
		r.Metadata = &struct{}{}
	case extsvc.TypePythonPackages:
		r.Metadata = &struct{}{}
	case extsvc.TypeRustPackages:
		r.Metadata = &struct{}{}
	default:
		log15.Warn("scanRepo - unknown service type", "type", typ)
		return nil
	}

	if err = json.Unmarshal(metadata, r.Metadata); err != nil {
		return errors.Wrapf(err, "scanRepo: failed to unmarshal %q metadata", typ)
	}

	return nil
}

// ReposListOptions specifies the options for listing repositories.
//
// Query and IncludePatterns/ExcludePatterns may not be used together.
type ReposListOptions struct {
	// What to select of each row.
	Select []string

	// Query specifies a search query for repositories. If specified, then the Sort and
	// Direction options are ignored
	Query string

	// IncludePatterns is a list of regular expressions, all of which must match all
	// repositories returned in the list.
	IncludePatterns []string

	// ExcludePattern is a regular expression that must not match any repository
	// returned in the list.
	ExcludePattern string

	// CaseSensitivePatterns determines if IncludePatterns and ExcludePattern are treated
	// with case sensitivity or not.
	CaseSensitivePatterns bool

	// Names is a list of repository names used to limit the results to that
	// set of repositories.
	// Note: This is currently used for version contexts. In future iterations,
	// version contexts may have their own table
	// and this may be replaced by the version context name.
	Names []string

	// HashedName is a repository hashed name used to limit the results to that repository.
	HashedName string

	// URIs selects any repos in the given set of URIs (i.e. uri column)
	URIs []string

	// IDs of repos to list. When zero-valued, this is omitted from the predicate set.
	IDs []api.RepoID

	// UserID, if non zero, will limit the set of results to repositories added by the user
	// through external services. Mutually exclusive with the ExternalServiceIDs and SearchContextID options.
	UserID int32

	// OrgID, if non zero, will limit the set of results to repositories owned by the organization
	// through external services. Mutually exclusive with the ExternalServiceIDs and SearchContextID options.
	OrgID int32

	// SearchContextID, if non zero, will limit the set of results to repositories listed in
	// the search context.
	SearchContextID int64

	// ExternalServiceIDs, if non empty, will only return repos added by the given external services.
	// The id is that of the external_services table NOT the external_service_id in the repo table
	// Mutually exclusive with the UserID option.
	ExternalServiceIDs []int64

	// ExternalRepos of repos to list. When zero-valued, this is omitted from the predicate set.
	ExternalRepos []api.ExternalRepoSpec

	// ExternalRepoIncludeContains is the list of specs to include repos using
	// SIMILAR TO matching. When zero-valued, this is omitted from the predicate set.
	ExternalRepoIncludeContains []api.ExternalRepoSpec

	// ExternalRepoExcludeContains is the list of specs to exclude repos using
	// SIMILAR TO matching. When zero-valued, this is omitted from the predicate set.
	ExternalRepoExcludeContains []api.ExternalRepoSpec

	// NoForks excludes forks from the list.
	NoForks bool

	// OnlyForks excludes non-forks from the lhist.
	OnlyForks bool

	// NoArchived excludes archived repositories from the list.
	NoArchived bool

	// OnlyArchived excludes non-archived repositories from the list.
	OnlyArchived bool

	// NoCloned excludes cloned repositories from the list.
	NoCloned bool

	// OnlyCloned excludes non-cloned repositories from the list.
	OnlyCloned bool

	// NoPrivate excludes private repositories from the list.
	NoPrivate bool

	// OnlyPrivate excludes non-private repositories from the list.
	OnlyPrivate bool

	// Index when set will only include repositories which should be indexed
	// if true. If false it will exclude repositories which should be
	// indexed. An example use case of this is for indexed search only
	// indexing a subset of repositories.
	Index *bool

	// List of fields by which to order the return repositories.
	OrderBy RepoListOrderBy

	// Cursors to efficiently paginate through large result sets.
	Cursors types.MultiCursor

	// UseOr decides between ANDing or ORing the predicates together.
	UseOr bool

	// IncludeUserPublicRepos will include repos from the user_public_repos table if this field is true, and the user_id
	// is non-zero. Note that these are not repos owned by this user, just ones they are interested in.
	IncludeUserPublicRepos bool

	// FailedFetch, if true, will filter to only repos that failed to clone or fetch
	// when last attempted. Specifically, this means that they have a non-null
	// last_error value in the gitserver_repos table.
	FailedFetch bool

	// MinLastChanged finds repository metadata or data that has changed since
	// MinLastChanged. It filters against repos.UpdatedAt,
	// gitserver.LastChanged and searchcontexts.UpdatedAt.
	//
	// LastChanged is the time of the last git fetch which changed refs
	// stored. IE the last time any branch changed (not just HEAD).
	//
	// UpdatedAt is the last time the metadata changed for a repository.
	//
	// Note: This option is used by our search indexer to determine what has
	// changed since it last polled. The fields its checks are all based on
	// what can affect search indexes.
	MinLastChanged time.Time

	// IncludeBlocked, if true, will include blocked repositories in the result set. Repos can be blocked
	// automatically or manually for different reasons, like being too big or having copyright issues.
	IncludeBlocked bool

	// IncludeDeleted, if true, will include soft deleted repositories in the result set.
	IncludeDeleted bool

	// joinGitserverRepos, if true, will make the fields of gitserver_repos available to select against,
	// with the table alias "gr".
	joinGitserverRepos bool

	// ExcludeSources, if true, will NULL out the Sources field on repo. Computing it is relatively costly
	// and if it doesn't end up being used this is wasted compute.
	ExcludeSources bool

	*LimitOffset
}

type RepoListOrderBy []RepoListSort

func (r RepoListOrderBy) SQL() *sqlf.Query {
	if len(r) == 0 {
		return sqlf.Sprintf("")
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
	Nulls      string
}

func (r RepoListSort) SQL() *sqlf.Query {
	var sb strings.Builder

	sb.WriteString(string(r.Field))

	if r.Descending {
		sb.WriteString(" DESC")
	}

	if r.Nulls == "FIRST" || r.Nulls == "LAST" {
		sb.WriteString(" NULLS " + r.Nulls)
	}

	return sqlf.Sprintf(sb.String())
}

// RepoListColumn is a column by which repositories can be sorted. These correspond to columns in the database.
type RepoListColumn string

const (
	RepoListCreatedAt RepoListColumn = "created_at"
	RepoListName      RepoListColumn = "name"
	RepoListID        RepoListColumn = "id"
	RepoListStars     RepoListColumn = "stars"
)

// List lists repositories in the Sourcegraph repository
//
// This will not return any repositories from external services that are not present in the Sourcegraph repository.
// Matching is done with fuzzy matching, i.e. "query" will match any repo name that matches the regexp `q.*u.*e.*r.*y`
func (s *repoStore) List(ctx context.Context, opt ReposListOptions) (results []*types.Repo, err error) {
	tr, ctx := trace.New(ctx, "repos.List", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// always having ID in ORDER BY helps Postgres create a more performant query plan
	if len(opt.OrderBy) == 0 || (len(opt.OrderBy) == 1 && opt.OrderBy[0].Field != RepoListID) {
		opt.OrderBy = append(opt.OrderBy, RepoListSort{Field: RepoListID})
	}

	return s.listRepos(ctx, tr, opt)
}

// StreamMinimalRepos calls the given callback for each of the repositories names and ids that match the given options.
func (s *repoStore) StreamMinimalRepos(ctx context.Context, opt ReposListOptions, cb func(*types.MinimalRepo)) (err error) {
	tr, ctx := trace.New(ctx, "repos.StreamMinimalRepos", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	opt.Select = minimalRepoColumns
	if len(opt.OrderBy) == 0 {
		opt.OrderBy = append(opt.OrderBy, RepoListSort{Field: RepoListID})
	}

	var privateIDs []api.RepoID

	err = s.list(ctx, tr, opt, func(rows *sql.Rows) error {
		var r types.MinimalRepo
		var private bool
		err := rows.Scan(&r.ID, &r.Name, &private, &dbutil.NullInt{N: &r.Stars})
		if err != nil {
			return err
		}

		cb(&r)

		if private {
			privateIDs = append(privateIDs, r.ID)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(privateIDs) > 0 {
		counterAccessGranted.Inc()
		logPrivateRepoAccessGranted(ctx, NewDBWith(s), privateIDs)
	}

	return nil
}

// ListMinimalRepos returns a list of repositories names and ids.
func (s *repoStore) ListMinimalRepos(ctx context.Context, opt ReposListOptions) (results []types.MinimalRepo, err error) {
	return results, s.StreamMinimalRepos(ctx, opt, func(r *types.MinimalRepo) {
		results = append(results, *r)
	})
}

func (s *repoStore) listRepos(ctx context.Context, tr *trace.Trace, opt ReposListOptions) (rs []*types.Repo, err error) {
	var privateIDs []api.RepoID
	err = s.list(ctx, tr, opt, func(rows *sql.Rows) error {
		var r types.Repo
		if err := scanRepo(rows, &r); err != nil {
			return err
		}

		rs = append(rs, &r)
		if r.Private {
			privateIDs = append(privateIDs, r.ID)
		}

		return nil
	})

	if len(privateIDs) > 0 {
		counterAccessGranted.Inc()
		logPrivateRepoAccessGranted(ctx, NewDBWith(s), privateIDs)
	}

	return rs, err
}

func (s *repoStore) list(ctx context.Context, tr *trace.Trace, opt ReposListOptions, scanRepo func(rows *sql.Rows) error) error {
	q, err := s.listSQL(ctx, opt)
	if err != nil {
		return err
	}

	tr.LogFields(trace.SQL(q))

	rows, err := s.Query(ctx, q)
	if err != nil {
		if e, ok := err.(*net.OpError); ok && e.Timeout() {
			return errors.Wrapf(context.DeadlineExceeded, "RepoStore.list: %s", err.Error())
		}
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := scanRepo(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (s *repoStore) listSQL(ctx context.Context, opt ReposListOptions) (*sqlf.Query, error) {
	var ctes, joins, where []*sqlf.Query

	// Cursor-based pagination requires parsing a handful of extra fields, which
	// may result in additional query conditions.
	if len(opt.Cursors) > 0 {
		cursorConds, err := parseCursorConds(opt.Cursors)
		if err != nil {
			return nil, err
		}

		if cursorConds != nil {
			where = append(where, cursorConds)
		}
	}

	if opt.Query != "" && (len(opt.IncludePatterns) > 0 || opt.ExcludePattern != "") {
		return nil, errors.New("Repos.List: Query and IncludePatterns/ExcludePattern options are mutually exclusive")
	}

	if opt.Query != "" {
		where = append(where, sqlf.Sprintf("lower(name) LIKE %s", "%"+strings.ToLower(opt.Query)+"%"))
	}

	for _, includePattern := range opt.IncludePatterns {
		extraConds, err := parsePattern(includePattern, opt.CaseSensitivePatterns)
		if err != nil {
			return nil, err
		}
		where = append(where, extraConds...)
	}

	if opt.ExcludePattern != "" {
		if opt.CaseSensitivePatterns {
			where = append(where, sqlf.Sprintf("name !~* %s", opt.ExcludePattern))
		} else {
			where = append(where, sqlf.Sprintf("lower(name) !~* %s", opt.ExcludePattern))
		}
	}

	if len(opt.IDs) > 0 {
		where = append(where, sqlf.Sprintf("id = ANY (%s)", pq.Array(opt.IDs)))
	}

	if len(opt.ExternalRepos) > 0 {
		er := make([]*sqlf.Query, 0, len(opt.ExternalRepos))
		for _, spec := range opt.ExternalRepos {
			er = append(er, sqlf.Sprintf("(external_id = %s AND external_service_type = %s AND external_service_id = %s)", spec.ID, spec.ServiceType, spec.ServiceID))
		}
		where = append(where, sqlf.Sprintf("(%s)", sqlf.Join(er, "\n OR ")))
	}

	if len(opt.ExternalRepoIncludeContains) > 0 {
		er := make([]*sqlf.Query, 0, len(opt.ExternalRepoIncludeContains))
		for _, spec := range opt.ExternalRepoIncludeContains {
			er = append(er, sqlf.Sprintf("(external_id SIMILAR TO %s AND external_service_type = %s AND external_service_id = %s)", spec.ID, spec.ServiceType, spec.ServiceID))
		}
		where = append(where, sqlf.Sprintf("(%s)", sqlf.Join(er, "\n OR ")))
	}

	if len(opt.ExternalRepoExcludeContains) > 0 {
		er := make([]*sqlf.Query, 0, len(opt.ExternalRepoExcludeContains))
		for _, spec := range opt.ExternalRepoExcludeContains {
			er = append(er, sqlf.Sprintf("(external_id NOT SIMILAR TO %s AND external_service_type = %s AND external_service_id = %s)", spec.ID, spec.ServiceType, spec.ServiceID))
		}
		where = append(where, sqlf.Sprintf("(%s)", sqlf.Join(er, "\n AND ")))
	}

	if opt.NoForks {
		where = append(where, sqlf.Sprintf("NOT fork"))
	}
	if opt.OnlyForks {
		where = append(where, sqlf.Sprintf("fork"))
	}
	if opt.NoArchived {
		where = append(where, sqlf.Sprintf("NOT archived"))
	}
	if opt.OnlyArchived {
		where = append(where, sqlf.Sprintf("archived"))
	}
	// Since https://github.com/sourcegraph/sourcegraph/pull/35633 there is no need to do an anti-join
	// with gitserver_repos table (checking for such repos that are present in repo but absent in gitserver_repos
	// table) because repo table is strictly consistent with gitserver_repos table.
	if opt.NoCloned {
		where = append(where, sqlf.Sprintf("(gr.clone_status IN ('not_cloned', 'cloning'))"))
	}
	if opt.OnlyCloned {
		where = append(where, sqlf.Sprintf("gr.clone_status = 'cloned'"))
	}
	if opt.FailedFetch {
		where = append(where, sqlf.Sprintf("gr.last_error IS NOT NULL"))
	}
	if !opt.MinLastChanged.IsZero() {
		conds := []*sqlf.Query{
			sqlf.Sprintf("gr.last_changed >= %s", opt.MinLastChanged),
			sqlf.Sprintf("COALESCE(repo.updated_at, repo.created_at) >= %s", opt.MinLastChanged),
			sqlf.Sprintf("repo.id IN (SELECT scr.repo_id FROM search_context_repos scr LEFT JOIN search_contexts sc ON scr.search_context_id = sc.id WHERE sc.updated_at >= %s)", opt.MinLastChanged),
		}
		where = append(where, sqlf.Sprintf("(%s)", sqlf.Join(conds, " OR ")))
	}
	if opt.NoPrivate {
		where = append(where, sqlf.Sprintf("NOT private"))
	}
	if opt.OnlyPrivate {
		where = append(where, sqlf.Sprintf("private"))
	}

	if len(opt.Names) > 0 {
		lowerNames := make([]string, len(opt.Names))
		for i, name := range opt.Names {
			lowerNames[i] = strings.ToLower(name)
		}

		// Performance improvement
		//
		// Comparing JUST the name field will use the repo_name_unique index, which is
		// a unique btree index over the citext name field. This tends to be a VERY SLOW
		// comparison over a large table. We were seeing query plans growing linearly with
		// the size of the result set such that each unique index scan would take ~0.1ms.
		// This adds up as we regularly query 10k-40k repositories at a time.
		//
		// This condition instead forces the use of a btree index repo_name_idx defined over
		// (lower(name::text) COLLATE "C"). This is a MUCH faster comparison as it does not
		// need to fold the casing of either the input value nor the value in the index.

		where = append(where, sqlf.Sprintf(`lower(name::text) COLLATE "C" = ANY (%s::text[])`, pq.Array(lowerNames)))
	}

	if opt.HashedName != "" {
		// This will use the repo_hashed_name_idx
		where = append(where, sqlf.Sprintf(`sha256(lower(name)::bytea) = decode(%s, 'hex')`, opt.HashedName))
	}

	if len(opt.URIs) > 0 {
		where = append(where, sqlf.Sprintf("uri = ANY (%s)", pq.Array(opt.URIs)))
	}

	if opt.Index != nil {
		// We don't currently have an index column, but when we want the
		// indexable repositories to be a subset it will live in the database
		// layer. So we do the filtering here.
		indexAll := conf.SearchIndexEnabled()
		if indexAll != *opt.Index {
			where = append(where, sqlf.Sprintf("false"))
		}
	}

	if (len(opt.ExternalServiceIDs) != 0 && (opt.UserID != 0 || opt.OrgID != 0)) ||
		(opt.UserID != 0 && opt.OrgID != 0) {
		return nil, errors.New("options ExternalServiceIDs, UserID and OrgID are mutually exclusive")
	} else if len(opt.ExternalServiceIDs) != 0 {
		where = append(where, sqlf.Sprintf("EXISTS (SELECT 1 FROM external_service_repos esr WHERE repo.id = esr.repo_id AND esr.external_service_id = ANY (%s))", pq.Array(opt.ExternalServiceIDs)))
	} else if opt.SearchContextID != 0 {
		// Joining on distinct search context repos to avoid returning duplicates
		joins = append(joins, sqlf.Sprintf(`JOIN (SELECT DISTINCT repo_id, search_context_id FROM search_context_repos) dscr ON repo.id = dscr.repo_id`))
		where = append(where, sqlf.Sprintf("dscr.search_context_id = %d", opt.SearchContextID))
	} else if opt.UserID != 0 {
		userReposCTE := sqlf.Sprintf(userReposCTEFmtstr, opt.UserID)
		if opt.IncludeUserPublicRepos {
			userReposCTE = sqlf.Sprintf("%s UNION %s", userReposCTE, sqlf.Sprintf(userPublicReposCTEFmtstr, opt.UserID))
		}
		ctes = append(ctes, sqlf.Sprintf("user_repos AS (%s)", userReposCTE))
		joins = append(joins, sqlf.Sprintf("JOIN user_repos ON user_repos.id = repo.id"))
	} else if opt.OrgID != 0 {
		joins = append(joins, sqlf.Sprintf("INNER JOIN external_service_repos ON external_service_repos.repo_id = repo.id"))
		where = append(where, sqlf.Sprintf("external_service_repos.org_id = %d", opt.OrgID))
	}

	if opt.NoCloned || opt.OnlyCloned || opt.FailedFetch || !opt.MinLastChanged.IsZero() || opt.joinGitserverRepos {
		joins = append(joins, sqlf.Sprintf("JOIN gitserver_repos gr ON gr.repo_id = repo.id"))
	}

	baseConds := sqlf.Sprintf("TRUE")
	if !opt.IncludeDeleted {
		baseConds = sqlf.Sprintf("repo.deleted_at IS NULL")
	}
	if !opt.IncludeBlocked {
		baseConds = sqlf.Sprintf("%s AND repo.blocked IS NULL", baseConds)
	}

	whereConds := sqlf.Sprintf("TRUE")
	if len(where) > 0 {
		if opt.UseOr {
			whereConds = sqlf.Join(where, "\n OR ")
		} else {
			whereConds = sqlf.Join(where, "\n AND ")
		}
	}

	queryConds := sqlf.Sprintf("%s AND (%s)", baseConds, whereConds)

	queryPrefix := sqlf.Sprintf("")
	if len(ctes) > 0 {
		queryPrefix = sqlf.Sprintf("WITH %s", sqlf.Join(ctes, ",\n"))
	}

	querySuffix := sqlf.Sprintf("%s %s", opt.OrderBy.SQL(), opt.LimitOffset.SQL())

	columns := repoColumns
	if !opt.ExcludeSources {
		columns = append(columns, getSourcesByRepoQueryStr)
	} else {
		columns = append(columns, "NULL")
	}
	if len(opt.Select) > 0 {
		columns = opt.Select
	}

	authzConds, err := AuthzQueryConds(ctx, NewDBWith(s))
	if err != nil {
		return nil, err
	}

	q := sqlf.Sprintf(
		fmt.Sprintf(listReposQueryFmtstr, strings.Join(columns, ",")),
		queryPrefix,
		sqlf.Join(joins, "\n"),
		queryConds,
		authzConds, // 🚨 SECURITY: Enforce repository permissions
		querySuffix,
	)

	return q, nil
}

const userReposCTEFmtstr = `
SELECT repo_id as id FROM external_service_repos WHERE user_id = %d
`

const userPublicReposCTEFmtstr = `
SELECT repo_id as id FROM user_public_repos WHERE user_id = %d
`

type ListIndexableReposOptions struct {
	// If true, will only include uncloned indexable repos
	OnlyUncloned bool
	// If true, we include user added private repos
	IncludePrivate bool

	*LimitOffset
}

var listIndexableReposMinStars, _ = strconv.Atoi(env.Get(
	"SRC_INDEXABLE_REPOS_MIN_STARS",
	"8",
	"Minimum stars needed for a public repo to be indexed on sourcegraph.com",
))

// ListIndexableRepos returns a list of repos to be indexed for search on sourcegraph.com.
// This includes all repos with >= SRC_INDEXABLE_REPOS_MIN_STARS stars as well as user or org added repos.
func (s *repoStore) ListIndexableRepos(ctx context.Context, opts ListIndexableReposOptions) (results []types.MinimalRepo, err error) {
	tr, ctx := trace.New(ctx, "repos.ListIndexable", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	var where, joins []*sqlf.Query

	if opts.OnlyUncloned {
		joins = append(joins, sqlf.Sprintf(
			"LEFT JOIN gitserver_repos gr ON gr.repo_id = repo.id",
		))
		where = append(where, sqlf.Sprintf(
			"(gr.clone_status IS NULL OR gr.clone_status = %s)",
			types.CloneStatusNotCloned,
		))
	}

	if !opts.IncludePrivate {
		where = append(where, sqlf.Sprintf("NOT repo.private"))
	}

	if len(where) == 0 {
		where = append(where, sqlf.Sprintf("TRUE"))
	}

	minStars := listIndexableReposMinStars
	if minStars == 0 {
		minStars = 8
	}

	q := sqlf.Sprintf(
		listIndexableReposQuery,
		sqlf.Join(joins, "\n"),
		minStars,
		sqlf.Join(where, "\nAND"),
		opts.LimitOffset.SQL(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "querying indexable repos")
	}
	defer rows.Close()

	for rows.Next() {
		var r types.MinimalRepo
		if err := rows.Scan(&r.ID, &r.Name, &dbutil.NullInt{N: &r.Stars}); err != nil {
			return nil, errors.Wrap(err, "scanning indexable repos")
		}
		results = append(results, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "scanning indexable repos")
	}

	return results, nil
}

const listIndexableReposQuery = `
-- source: internal/database/repos.go:ListIndexableRepos
SELECT
	repo.id, repo.name, repo.stars
FROM repo
%s
WHERE
	(
		(repo.stars >= %s AND NOT COALESCE(fork, false) AND NOT archived)
		OR
		lower(repo.name) ~ '^(src\.fedoraproject\.org|maven|npm|jdk)'
		OR
		repo.id IN (
			SELECT
				repo_id
			FROM
				external_service_repos
			WHERE
				external_service_repos.user_id IS NOT NULL
				OR
				external_service_repos.org_id IS NOT NULL

			UNION ALL

			SELECT
				repo_id
			FROM
				user_public_repos
		)
	)
	AND
	deleted_at IS NULL
	AND
	blocked IS NULL
	AND
	%s
ORDER BY stars DESC NULLS LAST
%s
`

// Create inserts repos and their sources, respectively in the repo and external_service_repos table.
// Associated external services must already exist.
func (s *repoStore) Create(ctx context.Context, repos ...*types.Repo) (err error) {
	tr, ctx := trace.New(ctx, "repos.Create", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	records := make([]*repoRecord, 0, len(repos))

	for _, r := range repos {
		repoRec, err := newRepoRecord(r)
		if err != nil {
			return err
		}

		records = append(records, repoRec)
	}

	encodedRepos, err := json.Marshal(records)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(insertReposQuery, string(encodedRepos))

	rows, err := s.Query(ctx, q)
	if err != nil {
		return errors.Wrap(err, "insert")
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for i := 0; rows.Next(); i++ {
		if err := rows.Scan(&repos[i].ID); err != nil {
			return err
		}
	}

	return nil
}

// repoRecord is the json representation of a repository as used in this package
// Postgres CTEs.
type repoRecord struct {
	ID                  api.RepoID      `json:"id"`
	Name                string          `json:"name"`
	URI                 *string         `json:"uri,omitempty"`
	Description         string          `json:"description"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           *time.Time      `json:"updated_at,omitempty"`
	DeletedAt           *time.Time      `json:"deleted_at,omitempty"`
	ExternalServiceType *string         `json:"external_service_type,omitempty"`
	ExternalServiceID   *string         `json:"external_service_id,omitempty"`
	ExternalID          *string         `json:"external_id,omitempty"`
	Archived            bool            `json:"archived"`
	Fork                bool            `json:"fork"`
	Stars               int             `json:"stars"`
	Private             bool            `json:"private"`
	Metadata            json.RawMessage `json:"metadata"`
	Sources             json.RawMessage `json:"sources,omitempty"`
}

func newRepoRecord(r *types.Repo) (*repoRecord, error) {
	metadata, err := metadataColumn(r.Metadata)
	if err != nil {
		return nil, errors.Wrapf(err, "newRecord: metadata marshalling failed")
	}

	sources, err := sourcesColumn(r.ID, r.Sources)
	if err != nil {
		return nil, errors.Wrapf(err, "newRecord: sources marshalling failed")
	}

	return &repoRecord{
		ID:                  r.ID,
		Name:                string(r.Name),
		URI:                 nullStringColumn(r.URI),
		Description:         r.Description,
		CreatedAt:           r.CreatedAt.UTC(),
		UpdatedAt:           nullTimeColumn(r.UpdatedAt),
		DeletedAt:           nullTimeColumn(r.DeletedAt),
		ExternalServiceType: nullStringColumn(r.ExternalRepo.ServiceType),
		ExternalServiceID:   nullStringColumn(r.ExternalRepo.ServiceID),
		ExternalID:          nullStringColumn(r.ExternalRepo.ID),
		Archived:            r.Archived,
		Fork:                r.Fork,
		Stars:               r.Stars,
		Private:             r.Private,
		Metadata:            metadata,
		Sources:             sources,
	}, nil
}

func nullTimeColumn(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func nullInt32Column(n int32) *int32 {
	if n == 0 {
		return nil
	}
	return &n
}

func nullStringColumn(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func metadataColumn(metadata any) (msg json.RawMessage, err error) {
	switch m := metadata.(type) {
	case nil:
		msg = json.RawMessage("{}")
	case string:
		msg = json.RawMessage(m)
	case []byte:
		msg = m
	case json.RawMessage:
		msg = m
	default:
		msg, err = json.MarshalIndent(m, "        ", "    ")
	}
	return
}

func sourcesColumn(repoID api.RepoID, sources map[string]*types.SourceInfo) (json.RawMessage, error) {
	var records []externalServiceRepo
	for _, src := range sources {
		records = append(records, externalServiceRepo{
			ExternalServiceID: src.ExternalServiceID(),
			RepoID:            int64(repoID),
			CloneURL:          src.CloneURL,
		})
	}

	return json.MarshalIndent(records, "        ", "    ")
}

type externalServiceRepo struct {
	ExternalServiceID int64  `json:"external_service_id"`
	RepoID            int64  `json:"repo_id"`
	CloneURL          string `json:"clone_url"`
}

var insertReposQuery = `
WITH repos_list AS (
  SELECT * FROM ROWS FROM (
	json_to_recordset(%s)
	AS (
		name                  citext,
		uri                   citext,
		description           text,
		created_at            timestamptz,
		updated_at            timestamptz,
		deleted_at            timestamptz,
		external_service_type text,
		external_service_id   text,
		external_id           text,
		archived              boolean,
		fork                  boolean,
		stars                 integer,
		private               boolean,
		metadata              jsonb,
		sources               jsonb
	  )
	)
	WITH ORDINALITY
),
inserted_repos AS (
  INSERT INTO repo (
	name,
	uri,
	description,
	created_at,
	updated_at,
	deleted_at,
	external_service_type,
	external_service_id,
	external_id,
	archived,
	fork,
	stars,
	private,
	metadata
  )
  SELECT
	name,
	NULLIF(BTRIM(uri), ''),
	description,
	created_at,
	updated_at,
	deleted_at,
	external_service_type,
	external_service_id,
	external_id,
	archived,
	fork,
	stars,
	private,
	metadata
  FROM repos_list
  RETURNING id
),
inserted_repos_rows AS (
  SELECT id, ROW_NUMBER() OVER () AS rn FROM inserted_repos
),
repos_list_rows AS (
  SELECT *, ROW_NUMBER() OVER () AS rn FROM repos_list
),
inserted_repos_with_ids AS (
  SELECT
	inserted_repos_rows.id,
	repos_list_rows.*
  FROM repos_list_rows
  JOIN inserted_repos_rows USING (rn)
),
sources_list AS (
  SELECT
    inserted_repos_with_ids.id AS repo_id,
	sources.external_service_id AS external_service_id,
	sources.clone_url AS clone_url
  FROM
    inserted_repos_with_ids,
	jsonb_to_recordset(inserted_repos_with_ids.sources)
	  AS sources(
		external_service_id bigint,
		repo_id             integer,
		clone_url           text
	  )
),
insert_sources AS (
  INSERT INTO external_service_repos (
    external_service_id,
    repo_id,
    user_id,
    org_id,
    clone_url
  )
  SELECT
    external_service_id,
    repo_id,
    es.namespace_user_id,
    es.namespace_org_id,
    clone_url
  FROM sources_list
  JOIN external_services es ON (es.id = external_service_id)
  ON CONFLICT ON CONSTRAINT external_service_repos_repo_id_external_service_id_unique
  DO
    UPDATE SET clone_url = EXCLUDED.clone_url
    WHERE external_service_repos.clone_url != EXCLUDED.clone_url
)
SELECT id FROM inserted_repos_with_ids;
`

// Delete deletes repos associated with the given ids and their associated sources.
func (s *repoStore) Delete(ctx context.Context, ids ...api.RepoID) error {
	if len(ids) == 0 {
		return nil
	}

	// The number of deleted repos can potentially be higher
	// than the maximum number of arguments we can pass to postgres.
	// We pass them as a json array instead to overcome this limitation.
	encodedIds, err := json.Marshal(ids)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(deleteReposQuery, string(encodedIds))

	err = s.Exec(ctx, q)
	if err != nil {
		return errors.Wrap(err, "delete")
	}

	return nil
}

const deleteReposQuery = `
WITH repo_ids AS (
  SELECT jsonb_array_elements_text(%s) AS id
)
UPDATE repo
SET
  name = soft_deleted_repository_name(name),
  deleted_at = transaction_timestamp()
FROM repo_ids
WHERE deleted_at IS NULL
AND repo.id = repo_ids.id::int
`

const getFirstRepoNamesByCloneURLQueryFmtstr = `
-- source:internal/database/repos.go:GetFirstRepoNameByCloneURL
SELECT
	name
FROM
	repo r
JOIN
	external_service_repos esr ON r.id = esr.repo_id
WHERE
	esr.clone_url = %s
ORDER BY
	r.updated_at DESC
LIMIT 1
`

// GetFirstRepoNameByCloneURL returns the first repo name in our database that
// match the given clone url. If not repo is found, an empty string and nil error
// are returned.
func (s *repoStore) GetFirstRepoNameByCloneURL(ctx context.Context, cloneURL string) (api.RepoName, error) {
	name, _, err := basestore.ScanFirstString(s.Query(ctx, sqlf.Sprintf(getFirstRepoNamesByCloneURLQueryFmtstr, cloneURL)))
	if err != nil {
		return "", err
	}
	return api.RepoName(name), nil
}

func parsePattern(p string, caseSensitive bool) ([]*sqlf.Query, error) {
	exact, like, pattern, err := parseIncludePattern(p)
	if err != nil {
		return nil, err
	}
	var conds []*sqlf.Query
	if exact != nil {
		if len(exact) == 0 || (len(exact) == 1 && exact[0] == "") {
			conds = append(conds, sqlf.Sprintf("TRUE"))
		} else {
			conds = append(conds, sqlf.Sprintf("name = ANY (%s)", pq.Array(exact)))
		}
	}
	for _, v := range like {
		if caseSensitive {
			conds = append(conds, sqlf.Sprintf(`name::text LIKE %s`, v))
		} else {
			conds = append(conds, sqlf.Sprintf(`lower(name) LIKE %s`, strings.ToLower(v)))
		}
	}
	if pattern != "" {
		if caseSensitive {
			conds = append(conds, sqlf.Sprintf("name::text ~ %s", pattern))
		} else {
			conds = append(conds, sqlf.Sprintf("lower(name) ~ lower(%s)", pattern))
		}
	}
	return []*sqlf.Query{sqlf.Sprintf("(%s)", sqlf.Join(conds, "OR"))}, nil
}

// parseCursorConds returns the WHERE conditions for the given cursor
func parseCursorConds(cs types.MultiCursor) (cond *sqlf.Query, err error) {
	var (
		direction string
		operator  string
		columns   = make([]string, 0, len(cs))
		values    = make([]*sqlf.Query, 0, len(cs))
	)

	for _, c := range cs {
		if c == nil || c.Column == "" || c.Value == "" {
			continue
		}

		if direction == "" {
			switch direction = c.Direction; direction {
			case "next":
				operator = ">="
			case "prev":
				operator = "<="
			default:
				return nil, errors.Errorf("missing or invalid cursor direction: %q", c.Direction)
			}
		} else if direction != c.Direction {
			return nil, errors.Errorf("multi-cursors must have the same direction")
		}

		switch RepoListColumn(c.Column) {
		case RepoListName, RepoListStars, RepoListCreatedAt, RepoListID:
			columns = append(columns, c.Column)
			values = append(values, sqlf.Sprintf("%s", c.Value))
		default:
			return nil, errors.Errorf("missing or invalid cursor: %q %q", c.Column, c.Value)
		}
	}

	if len(columns) == 0 {
		return nil, nil
	}

	return sqlf.Sprintf(fmt.Sprintf("(%s) %s (%%s)", strings.Join(columns, ", "), operator), sqlf.Join(values, ", ")), nil
}

// parseIncludePattern either (1) parses the pattern into a list of exact possible
// string values and LIKE patterns if such a list can be determined from the pattern,
// and (2) returns the original regexp if those patterns are not equivalent to the
// regexp.
//
// It allows Repos.List to optimize for the common case where a pattern like
// `(^github.com/foo/bar$)|(^github.com/baz/qux$)` is provided. In that case,
// it's faster to query for "WHERE name IN (...)" the two possible exact values
// (because it can use an index) instead of using a "WHERE name ~*" regexp condition
// (which generally can't use an index).
//
// This optimization is necessary for good performance when there are many repos
// in the database. With this optimization, specifying a "repogroup:" in the query
// will be fast (even if there are many repos) because the query can be constrained
// efficiently to only the repos in the group.
func parseIncludePattern(pattern string) (exact, like []string, regexp string, err error) {
	re, err := regexpsyntax.Parse(pattern, regexpsyntax.Perl)
	if err != nil {
		return nil, nil, "", err
	}
	exact, contains, prefix, suffix, err := allMatchingStrings(re.Simplify(), false)
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
// matches, if it's possible to determine the list. The "last" argument
// indicates if this is the last part of the original regexp.
func allMatchingStrings(re *regexpsyntax.Regexp, last bool) (exact, contains, prefix, suffix []string, err error) {
	switch re.Op {
	case regexpsyntax.OpEmptyMatch:
		return []string{""}, nil, nil, nil, nil
	case regexpsyntax.OpLiteral:
		prog, err := regexpsyntax.Compile(re)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		prefix, complete := prog.Prefix()
		if complete {
			return nil, []string{prefix}, nil, nil, nil
		}
		return nil, nil, nil, nil, nil

	case regexpsyntax.OpCharClass:
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

	case regexpsyntax.OpStar:
		if len(re.Sub) == 1 && (re.Sub[0].Op == regexpsyntax.OpAnyCharNotNL || re.Sub[0].Op == regexpsyntax.OpAnyChar) {
			if last {
				return nil, []string{""}, nil, nil, nil
			}
			return nil, nil, nil, nil, nil
		}

	case regexpsyntax.OpBeginText:
		return nil, nil, []string{""}, nil, nil

	case regexpsyntax.OpEndText:
		return nil, nil, nil, []string{""}, nil

	case regexpsyntax.OpCapture:
		return allMatchingStrings(re.Sub0[0], false)

	case regexpsyntax.OpConcat:
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
			subexact, subcontains, subprefix, subsuffix, err := allMatchingStrings(sub, i == len(re.Sub)-1)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			if subexact == nil && subcontains == nil && subprefix == nil && subsuffix == nil {
				return nil, nil, nil, nil, nil
			}

			// We only returns subcontains for child literals. But because it
			// is part of a concat pattern, we know it is exact when we
			// append. This transformation has been running in production for
			// many years, so while it isn't correct for all inputs
			// theoretically, in practice this hasn't been a problem. However,
			// a redesign of this function as a whole is needed. - keegan
			if subcontains != nil {
				subexact = append(subexact, subcontains...)
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

	case regexpsyntax.OpAlternate:
		for _, sub := range re.Sub {
			subexact, subcontains, subprefix, subsuffix, err := allMatchingStrings(sub, false)
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
