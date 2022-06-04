package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrSearchContextNotFound = errors.New("search context not found")

func SearchContextsWith(other basestore.ShareableStore) SearchContextsStore {
	return &searchContextsStore{basestore.NewWithHandle(other.Handle())}
}

type SearchContextsStore interface {
	basestore.ShareableStore
	CountSearchContexts(context.Context, ListSearchContextsOptions) (int32, error)
	CreateSearchContextWithRepositoryRevisions(context.Context, *types.SearchContext, []*types.SearchContextRepositoryRevisions) (*types.SearchContext, error)
	DeleteSearchContext(context.Context, int64) error
	Done(error) error
	Exec(context.Context, *sqlf.Query) error
	GetAllRevisionsForRepos(context.Context, []api.RepoID) (map[api.RepoID][]string, error)
	GetSearchContext(context.Context, GetSearchContextOptions) (*types.SearchContext, error)
	GetSearchContextRepositoryRevisions(context.Context, int64) ([]*types.SearchContextRepositoryRevisions, error)
	ListSearchContexts(context.Context, ListSearchContextsPageOptions, ListSearchContextsOptions) ([]*types.SearchContext, error)
	GetAllQueries(context.Context) ([]string, error)
	SetSearchContextRepositoryRevisions(context.Context, int64, []*types.SearchContextRepositoryRevisions) error
	Transact(context.Context) (SearchContextsStore, error)
	UpdateSearchContextWithRepositoryRevisions(context.Context, *types.SearchContext, []*types.SearchContextRepositoryRevisions) (*types.SearchContext, error)
}

type searchContextsStore struct {
	*basestore.Store
}

func (s *searchContextsStore) Transact(ctx context.Context) (SearchContextsStore, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &searchContextsStore{Store: txBase}, nil
}

const searchContextsPermissionsConditionFmtStr = `(
    -- Bypass permission check
    %s
    -- Happy path of public search contexts
    OR sc.public
    -- Private user contexts are available only to its creator
    OR (sc.namespace_user_id IS NOT NULL AND sc.namespace_user_id = %d)
    -- Private org contexts are available only to its members
    OR (sc.namespace_org_id IS NOT NULL AND EXISTS (SELECT FROM org_members om WHERE om.org_id = sc.namespace_org_id AND om.user_id = %d))
    -- Private instance-level contexts are available only to site-admins
    OR (sc.namespace_user_id IS NULL AND sc.namespace_org_id IS NULL AND EXISTS (SELECT FROM users u WHERE u.id = %d AND u.site_admin))
)`

func searchContextsPermissionsCondition(ctx context.Context, store basestore.ShareableStore) (*sqlf.Query, error) {
	a := actor.FromContext(ctx)
	authenticatedUserID := int32(0)
	bypassPermissionsCheck := a.Internal
	if !bypassPermissionsCheck && a.IsAuthenticated() {
		currentUser, err := UsersWith(store).GetByCurrentAuthUser(ctx)
		if err != nil {
			return nil, err
		}
		authenticatedUserID = currentUser.ID
	}
	q := sqlf.Sprintf(searchContextsPermissionsConditionFmtStr, bypassPermissionsCheck, authenticatedUserID, authenticatedUserID, authenticatedUserID)
	return q, nil
}

const listSearchContextsFmtStr = `
SELECT
  sc.id,
  sc.name,
  sc.description,
  sc.public,
  sc.namespace_user_id,
  sc.namespace_org_id,
  sc.updated_at,
  sc.query,
  u.username,
  o.name
FROM search_contexts sc
LEFT JOIN users u on sc.namespace_user_id = u.id
LEFT JOIN orgs o on sc.namespace_org_id = o.id
WHERE
	(%s) -- permission conditions
	AND (%s) -- query conditions
ORDER BY %s
LIMIT %d
OFFSET %d
`

const countSearchContextsFmtStr = `
SELECT COUNT(*)
FROM search_contexts sc
LEFT JOIN users u on sc.namespace_user_id = u.id
LEFT JOIN orgs o on sc.namespace_org_id = o.id
WHERE
	(%s) -- permission conditions
	AND (%s) -- query conditions
`

type SearchContextsOrderByOption uint8

const (
	SearchContextsOrderByID SearchContextsOrderByOption = iota
	SearchContextsOrderBySpec
	SearchContextsOrderByUpdatedAt
)

type ListSearchContextsPageOptions struct {
	First int32
	After int32
}

// ListSearchContextsOptions specifies the options for listing search contexts.
// It produces a union of all search contexts that match NamespaceUserIDs, or NamespaceOrgIDs, or NoNamespace. If none of those
// are specified, it produces all available search contexts.
type ListSearchContextsOptions struct {
	// Name is used for partial matching of search contexts by name (case-insensitvely).
	Name string
	// NamespaceName is used for partial matching of search context namespaces (user or org) by name (case-insensitvely).
	NamespaceName string
	// NamespaceUserIDs matches search contexts by user namespace. If multiple IDs are specified, then a union of all matching results is returned.
	NamespaceUserIDs []int32
	// NamespaceOrgIDs matches search contexts by org. If multiple IDs are specified, then a union of all matching results is returned.
	NamespaceOrgIDs []int32
	// NoNamespace matches search contexts without a namespace ("instance-level contexts").
	NoNamespace bool
	// OrderBy specifies the ordering option for search contexts. Search contexts are ordered using SearchContextsOrderByID by default.
	// SearchContextsOrderBySpec option sorts contexts by coallesced namespace names first
	// (user name and org name) and then by context name. SearchContextsOrderByUpdatedAt option sorts
	// search contexts by their last update time (updated_at).
	OrderBy SearchContextsOrderByOption
	// OrderByDescending specifies the sort direction for the OrderBy option.
	OrderByDescending bool
}

func getSearchContextOrderByClause(orderBy SearchContextsOrderByOption, descending bool) *sqlf.Query {
	orderDirection := "ASC"
	if descending {
		orderDirection = "DESC"
	}
	switch orderBy {
	case SearchContextsOrderBySpec:
		return sqlf.Sprintf(fmt.Sprintf("COALESCE(u.username, o.name) %s, sc.name %s", orderDirection, orderDirection))
	case SearchContextsOrderByUpdatedAt:
		return sqlf.Sprintf("sc.updated_at " + orderDirection)
	case SearchContextsOrderByID:
		return sqlf.Sprintf("sc.id " + orderDirection)
	}
	panic("invalid SearchContextsOrderByOption option")
}

func getSearchContextNamespaceQueryConditions(namespaceUserID, namespaceOrgID int32) ([]*sqlf.Query, error) {
	conds := []*sqlf.Query{}
	if namespaceUserID != 0 && namespaceOrgID != 0 {
		return nil, errors.New("options NamespaceUserID and NamespaceOrgID are mutually exclusive")
	}
	if namespaceUserID > 0 {
		conds = append(conds, sqlf.Sprintf("sc.namespace_user_id = %s", namespaceUserID))
	}
	if namespaceOrgID > 0 {
		conds = append(conds, sqlf.Sprintf("sc.namespace_org_id = %s", namespaceOrgID))
	}
	return conds, nil
}

func idsToQueries(ids []int32) []*sqlf.Query {
	queries := make([]*sqlf.Query, 0, len(ids))
	for _, id := range ids {
		queries = append(queries, sqlf.Sprintf("%s", id))
	}
	return queries
}

func getSearchContextsQueryConditions(opts ListSearchContextsOptions) []*sqlf.Query {
	namespaceConds := []*sqlf.Query{}
	if opts.NoNamespace {
		namespaceConds = append(namespaceConds, sqlf.Sprintf("(sc.namespace_user_id IS NULL AND sc.namespace_org_id IS NULL)"))
	}
	if len(opts.NamespaceUserIDs) > 0 {
		namespaceConds = append(namespaceConds, sqlf.Sprintf("sc.namespace_user_id IN (%s)", sqlf.Join(idsToQueries(opts.NamespaceUserIDs), ",")))
	}
	if len(opts.NamespaceOrgIDs) > 0 {
		namespaceConds = append(namespaceConds, sqlf.Sprintf("sc.namespace_org_id IN (%s)", sqlf.Join(idsToQueries(opts.NamespaceOrgIDs), ",")))
	}

	conds := []*sqlf.Query{}
	if len(namespaceConds) > 0 {
		conds = append(conds, sqlf.Sprintf("(%s)", sqlf.Join(namespaceConds, " OR ")))
	}

	if opts.Name != "" {
		// name column has type citext which automatically performs case-insensitive comparison
		conds = append(conds, sqlf.Sprintf("sc.name LIKE %s", "%"+opts.Name+"%"))
	}

	if opts.NamespaceName != "" {
		conds = append(conds, sqlf.Sprintf("COALESCE(u.username, o.name, '') ILIKE %s", "%"+opts.NamespaceName+"%"))
	}

	if len(conds) == 0 {
		// If no conditions are present, append a catch-all condition to avoid a SQL syntax error
		conds = append(conds, sqlf.Sprintf("1 = 1"))
	}

	return conds
}

func (s *searchContextsStore) listSearchContexts(ctx context.Context, cond *sqlf.Query, orderBy *sqlf.Query, limit int32, offset int32) ([]*types.SearchContext, error) {
	permissionsCond, err := searchContextsPermissionsCondition(ctx, s)
	if err != nil {
		return nil, err
	}
	rows, err := s.Query(ctx, sqlf.Sprintf(listSearchContextsFmtStr, permissionsCond, cond, orderBy, limit, offset))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSearchContexts(rows)
}

func (s *searchContextsStore) ListSearchContexts(ctx context.Context, pageOpts ListSearchContextsPageOptions, opts ListSearchContextsOptions) ([]*types.SearchContext, error) {
	conds := getSearchContextsQueryConditions(opts)
	orderBy := getSearchContextOrderByClause(opts.OrderBy, opts.OrderByDescending)
	return s.listSearchContexts(ctx, sqlf.Join(conds, "\n AND "), orderBy, pageOpts.First, pageOpts.After)
}

func (s *searchContextsStore) CountSearchContexts(ctx context.Context, opts ListSearchContextsOptions) (int32, error) {
	conds := getSearchContextsQueryConditions(opts)
	permissionsCond, err := searchContextsPermissionsCondition(ctx, s)
	if err != nil {
		return -1, err
	}
	var count int32
	err = s.QueryRow(ctx, sqlf.Sprintf(countSearchContextsFmtStr, permissionsCond, sqlf.Join(conds, "\n AND "))).Scan(&count)
	if err != nil {
		return -1, err
	}
	return count, err
}

type GetSearchContextOptions struct {
	Name            string
	NamespaceUserID int32
	NamespaceOrgID  int32
}

func (s *searchContextsStore) GetSearchContext(ctx context.Context, opts GetSearchContextOptions) (*types.SearchContext, error) {
	conds := []*sqlf.Query{}
	if opts.NamespaceUserID == 0 && opts.NamespaceOrgID == 0 {
		conds = append(conds, sqlf.Sprintf("sc.namespace_user_id IS NULL"), sqlf.Sprintf("sc.namespace_org_id IS NULL"))
	} else {
		namespaceConds, err := getSearchContextNamespaceQueryConditions(opts.NamespaceUserID, opts.NamespaceOrgID)
		if err != nil {
			return nil, err
		}
		conds = append(conds, namespaceConds...)
	}
	conds = append(conds, sqlf.Sprintf("sc.name = %s", opts.Name))

	permissionsCond, err := searchContextsPermissionsCondition(ctx, s)
	if err != nil {
		return nil, err
	}
	rows, err := s.Query(
		ctx,
		sqlf.Sprintf(
			listSearchContextsFmtStr,
			permissionsCond,
			sqlf.Join(conds, "\n AND "),
			getSearchContextOrderByClause(SearchContextsOrderByID, false),
			1, // limit
			0, // offset
		),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSingleSearchContext(rows)
}

const deleteSearchContextFmtStr = `
DELETE FROM search_contexts WHERE id = %d
`

// 🚨 SECURITY: The caller must ensure that the actor is a site admin or has permission to delete the search context.
func (s *searchContextsStore) DeleteSearchContext(ctx context.Context, searchContextID int64) error {
	return s.Exec(ctx, sqlf.Sprintf(deleteSearchContextFmtStr, searchContextID))
}

const insertSearchContextFmtStr = `
INSERT INTO search_contexts
(name, description, public, namespace_user_id, namespace_org_id, query)
VALUES (%s, %s, %s, %s, %s, %s)
`

// 🚨 SECURITY: The caller must ensure that the actor is a site admin or has permission to create the search context.
func (s *searchContextsStore) CreateSearchContextWithRepositoryRevisions(ctx context.Context, searchContext *types.SearchContext, repositoryRevisions []*types.SearchContextRepositoryRevisions) (createdSearchContext *types.SearchContext, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	createdSearchContext, err = createSearchContext(ctx, tx, searchContext)
	if err != nil {
		return nil, err
	}

	err = tx.SetSearchContextRepositoryRevisions(ctx, createdSearchContext.ID, repositoryRevisions)
	if err != nil {
		return nil, err
	}
	return createdSearchContext, nil
}

const updateSearchContextFmtStr = `
UPDATE search_contexts
SET
	name = %s,
	description = %s,
	public = %s,
	query = %s,
	updated_at = now()
WHERE id = %d
`

// 🚨 SECURITY: The caller must ensure that the actor is a site admin or has permission to update the search context.
func (s *searchContextsStore) UpdateSearchContextWithRepositoryRevisions(ctx context.Context, searchContext *types.SearchContext, repositoryRevisions []*types.SearchContextRepositoryRevisions) (_ *types.SearchContext, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	updatedSearchContext, err := updateSearchContext(ctx, tx, searchContext)
	if err != nil {
		return nil, err
	}

	err = tx.SetSearchContextRepositoryRevisions(ctx, updatedSearchContext.ID, repositoryRevisions)
	if err != nil {
		return nil, err
	}
	return updatedSearchContext, nil
}

func (s *searchContextsStore) SetSearchContextRepositoryRevisions(ctx context.Context, searchContextID int64, repositoryRevisions []*types.SearchContextRepositoryRevisions) (err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	err = tx.Exec(ctx, sqlf.Sprintf("DELETE FROM search_context_repos WHERE search_context_id = %d", searchContextID))
	if err != nil {
		return err
	}

	if len(repositoryRevisions) == 0 {
		return nil
	}

	values := []*sqlf.Query{}
	for _, repoRev := range repositoryRevisions {
		for _, revision := range repoRev.Revisions {
			values = append(values, sqlf.Sprintf(
				"(%s, %s, %s)",
				searchContextID, repoRev.Repo.ID, revision,
			))
		}
	}

	return tx.Exec(ctx, sqlf.Sprintf(
		"INSERT INTO search_context_repos (search_context_id, repo_id, revision) VALUES %s",
		sqlf.Join(values, ","),
	))
}

func createSearchContext(ctx context.Context, s SearchContextsStore, searchContext *types.SearchContext) (*types.SearchContext, error) {
	q := sqlf.Sprintf(
		insertSearchContextFmtStr,
		searchContext.Name,
		searchContext.Description,
		searchContext.Public,
		nullInt32Column(searchContext.NamespaceUserID),
		nullInt32Column(searchContext.NamespaceOrgID),
		nullStringColumn(searchContext.Query),
	)
	_, err := s.Handle().DB().ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	return s.GetSearchContext(ctx, GetSearchContextOptions{
		Name:            searchContext.Name,
		NamespaceUserID: searchContext.NamespaceUserID,
		NamespaceOrgID:  searchContext.NamespaceOrgID,
	})
}

func updateSearchContext(ctx context.Context, s SearchContextsStore, searchContext *types.SearchContext) (*types.SearchContext, error) {
	q := sqlf.Sprintf(
		updateSearchContextFmtStr,
		searchContext.Name,
		searchContext.Description,
		searchContext.Public,
		nullStringColumn(searchContext.Query),
		searchContext.ID,
	)
	_, err := s.Handle().DB().ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	return s.GetSearchContext(ctx, GetSearchContextOptions{
		Name:            searchContext.Name,
		NamespaceUserID: searchContext.NamespaceUserID,
		NamespaceOrgID:  searchContext.NamespaceOrgID,
	})
}

func scanSingleSearchContext(rows *sql.Rows) (*types.SearchContext, error) {
	searchContexts, err := scanSearchContexts(rows)
	if err != nil {
		return nil, err
	}
	if len(searchContexts) != 1 {
		return nil, ErrSearchContextNotFound
	}
	return searchContexts[0], nil
}

func scanSearchContexts(rows *sql.Rows) ([]*types.SearchContext, error) {
	var out []*types.SearchContext
	for rows.Next() {
		sc := &types.SearchContext{}
		err := rows.Scan(
			&sc.ID,
			&sc.Name,
			&sc.Description,
			&sc.Public,
			&dbutil.NullInt32{N: &sc.NamespaceUserID},
			&dbutil.NullInt32{N: &sc.NamespaceOrgID},
			&sc.UpdatedAt,
			&dbutil.NullString{S: &sc.Query},
			&dbutil.NullString{S: &sc.NamespaceUserName},
			&dbutil.NullString{S: &sc.NamespaceOrgName},
		)
		if err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, nil
}

var getSearchContextRepositoryRevisionsFmtStr = `
-- source:internal/database/search_contexts.go:GetSearchContextRepositoryRevisions
SELECT
	sc.repo_id,
	sc.revision,
	r.name
FROM
	search_context_repos sc
JOIN
	(
		SELECT
			id,
			name
		FROM repo
		WHERE
			deleted_at IS NULL
			AND
			blocked IS NULL
			AND (%s) -- populates authzConds
	) r
	ON r.id = sc.repo_id
WHERE sc.search_context_id = %d
`

func (s *searchContextsStore) GetSearchContextRepositoryRevisions(ctx context.Context, searchContextID int64) ([]*types.SearchContextRepositoryRevisions, error) {
	authzConds, err := AuthzQueryConds(ctx, NewDB(s.Handle().DB()))
	if err != nil {
		return nil, err
	}

	rows, err := s.Query(ctx, sqlf.Sprintf(
		getSearchContextRepositoryRevisionsFmtStr,
		authzConds,
		searchContextID,
	))
	if err != nil {
		return nil, err
	}

	defer func() {
		err = basestore.CloseRows(rows, err)
	}()

	repositoryIDsToRevisions := map[int32][]string{}
	repositoryIDsToName := map[int32]string{}
	for rows.Next() {
		var repoID int32
		var repoName, revision string
		err = rows.Scan(&repoID, &revision, &repoName)
		if err != nil {
			return nil, err
		}
		repositoryIDsToRevisions[repoID] = append(repositoryIDsToRevisions[repoID], revision)
		repositoryIDsToName[repoID] = repoName
	}

	out := make([]*types.SearchContextRepositoryRevisions, 0, len(repositoryIDsToRevisions))
	for repoID, revisions := range repositoryIDsToRevisions {
		sort.Strings(revisions)

		out = append(out, &types.SearchContextRepositoryRevisions{
			Repo: types.MinimalRepo{
				ID:   api.RepoID(repoID),
				Name: api.RepoName(repositoryIDsToName[repoID]),
			},
			Revisions: revisions,
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Repo.ID < out[j].Repo.ID })

	return out, nil
}

var getAllRevisionsForReposFmtStr = `
-- source:internal/database/search_contexts.go:GetAllRevisionsForRepos
SELECT DISTINCT
	scr.repo_id,
	scr.revision
FROM
	search_context_repos scr
WHERE
	scr.repo_id = ANY (%s)
ORDER BY
	scr.revision
`

// GetAllRevisionsForRepos returns the list of revisions that are used in search
// contexts for each given repo ID.
func (s *searchContextsStore) GetAllRevisionsForRepos(ctx context.Context, repoIDs []api.RepoID) (map[api.RepoID][]string, error) {
	if a := actor.FromContext(ctx); !a.IsInternal() {
		return nil, errors.New("GetAllRevisionsForRepos can only be accessed by an internal actor")
	}

	if len(repoIDs) == 0 {
		return map[api.RepoID][]string{}, nil
	}

	q := sqlf.Sprintf(
		getAllRevisionsForReposFmtStr,
		pq.Array(repoIDs),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	revs := make(map[api.RepoID][]string, len(repoIDs))
	for rows.Next() {
		var (
			repoID api.RepoID
			rev    string
		)
		if err = rows.Scan(&repoID, &rev); err != nil {
			return nil, err
		}
		revs[repoID] = append(revs[repoID], rev)
	}

	return revs, nil
}

func (s *searchContextsStore) GetAllQueries(ctx context.Context) (qs []string, _ error) {
	if a := actor.FromContext(ctx); !a.IsInternal() {
		return nil, errors.New("GetAllQueries can only be accessed by an internal actor")
	}

	q := sqlf.Sprintf(`SELECT array_agg(query) FROM search_contexts WHERE query IS NOT NULL`)

	return qs, s.QueryRow(ctx, q).Scan(pq.Array(&qs))
}
