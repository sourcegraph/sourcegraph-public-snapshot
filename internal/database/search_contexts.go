package database

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func SearchContexts(db dbutil.DB) *SearchContextsStore {
	store := basestore.NewWithDB(db, sql.TxOptions{})
	return &SearchContextsStore{store}
}

type SearchContextsStore struct {
	*basestore.Store
}

func (s *SearchContextsStore) Transact(ctx context.Context) (*SearchContextsStore, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &SearchContextsStore{Store: txBase}, nil
}

const listSearchContextsFmtStr = `
SELECT id, name, description, public, namespace_user_id, namespace_org_id
FROM search_contexts
WHERE deleted_at IS NULL AND (%s)
`

func (s *SearchContextsStore) listSearchContexts(ctx context.Context, cond *sqlf.Query) ([]*types.SearchContext, error) {
	rows, err := s.Query(ctx, sqlf.Sprintf(listSearchContextsFmtStr, cond))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSearchContexts(rows)
}

func (s *SearchContextsStore) ListSearchContextsByUserID(ctx context.Context, userID int32) ([]*types.SearchContext, error) {
	if Mocks.SearchContexts.ListSearchContextsByUserID != nil {
		return Mocks.SearchContexts.ListSearchContextsByUserID(ctx, userID)
	}
	return s.listSearchContexts(ctx, sqlf.Sprintf("namespace_user_id = %d", userID))
}

func (s *SearchContextsStore) ListInstanceLevelSearchContexts(ctx context.Context) ([]*types.SearchContext, error) {
	if Mocks.SearchContexts.ListInstanceLevelSearchContexts != nil {
		return Mocks.SearchContexts.ListInstanceLevelSearchContexts(ctx)
	}
	return s.listSearchContexts(ctx, sqlf.Sprintf("namespace_user_id IS NULL AND namespace_org_id IS NULL"))
}

const getSearchContextFmtStr = listSearchContextsFmtStr + "\nLIMIT 1"

type GetSearchContextOptions struct {
	Name            string
	NamespaceUserID int32
	NamespaceOrgID  int32
}

func (s *SearchContextsStore) GetSearchContext(ctx context.Context, opts GetSearchContextOptions) (*types.SearchContext, error) {
	if Mocks.SearchContexts.GetSearchContext != nil {
		return Mocks.SearchContexts.GetSearchContext(ctx, opts)
	}

	conds := []*sqlf.Query{}

	if opts.NamespaceUserID != 0 && opts.NamespaceOrgID != 0 {
		return nil, errors.New("options NamespaceUserID and NamespaceOrgID are mutually exclusive")
	}

	if opts.NamespaceUserID == 0 {
		conds = append(conds, sqlf.Sprintf("namespace_user_id IS NULL"))
	} else {
		conds = append(conds, sqlf.Sprintf("namespace_user_id = %s", opts.NamespaceUserID))
	}
	if opts.NamespaceOrgID == 0 {
		conds = append(conds, sqlf.Sprintf("namespace_org_id IS NULL"))
	} else {
		conds = append(conds, sqlf.Sprintf("namespace_org_id = %s", opts.NamespaceOrgID))
	}
	if opts.Name != "" {
		conds = append(conds, sqlf.Sprintf("name = %s", opts.Name))
	}

	rows, err := s.Query(ctx, sqlf.Sprintf(getSearchContextFmtStr, sqlf.Join(conds, "\n AND ")))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSingleSearchContext(rows)
}

const insertSearchContextFmtStr = `
INSERT INTO search_contexts
(name, description, public, namespace_user_id, namespace_org_id)
VALUES (%s, %s, %s, %s, %s)
RETURNING id, name, description, public, namespace_user_id, namespace_org_id;
`

func (s *SearchContextsStore) CreateSearchContextWithRepositoryRevisions(ctx context.Context, searchContext *types.SearchContext, repositoryRevisions []*types.SearchContextRepositoryRevisions) (createdSearchContext *types.SearchContext, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	createdSearchContext, err = tx.createSearchContext(ctx, searchContext)
	if err != nil {
		return nil, err
	}

	err = tx.SetSearchContextRepositoryRevisions(ctx, createdSearchContext.ID, repositoryRevisions)
	if err != nil {
		return nil, err
	}
	return createdSearchContext, nil
}

func (s *SearchContextsStore) SetSearchContextRepositoryRevisions(ctx context.Context, searchContextID int64, repositoryRevisions []*types.SearchContextRepositoryRevisions) (err error) {
	if len(repositoryRevisions) == 0 {
		return nil
	}

	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	err = tx.Exec(ctx, sqlf.Sprintf("DELETE FROM search_context_repos WHERE search_context_id = %d", searchContextID))
	if err != nil {
		return err
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

func (s *SearchContextsStore) createSearchContext(ctx context.Context, searchContext *types.SearchContext) (*types.SearchContext, error) {
	rows, err := s.Query(ctx, sqlf.Sprintf(
		insertSearchContextFmtStr,
		searchContext.Name,
		searchContext.Description,
		// Always insert search context as public until private contexts are supported
		true,
		nullInt32Column(searchContext.NamespaceUserID),
		nullInt32Column(searchContext.NamespaceOrgID),
	))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSingleSearchContext(rows)
}

func scanSingleSearchContext(rows *sql.Rows) (*types.SearchContext, error) {
	searchContexts, err := scanSearchContexts(rows)
	if err != nil {
		return nil, err
	}
	if len(searchContexts) != 1 {
		return nil, errors.New("search context not found")
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
		)
		if err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, nil
}

var getSearchContextRepositoryRevisionsFmtStr = `
SELECT sc.repo_id, sc.revision, r.name
FROM search_context_repos sc
JOIN
	(SELECT id, name FROM repo WHERE deleted_at IS NULL AND (%s)) r -- populates authzConds
	ON r.id = sc.repo_id
WHERE sc.search_context_id = %d
`

func (s *SearchContextsStore) GetSearchContextRepositoryRevisions(ctx context.Context, searchContextID int64) ([]*types.SearchContextRepositoryRevisions, error) {
	if Mocks.SearchContexts.GetSearchContextRepositoryRevisions != nil {
		return Mocks.SearchContexts.GetSearchContextRepositoryRevisions(ctx, searchContextID)
	}

	authzConds, err := AuthzQueryConds(ctx, s.Handle().DB())
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
	defer rows.Close()

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
			Repo: &types.RepoName{
				ID:   api.RepoID(repoID),
				Name: api.RepoName(repositoryIDsToName[repoID]),
			},
			Revisions: revisions,
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Repo.ID < out[j].Repo.ID })
	return out, nil
}
