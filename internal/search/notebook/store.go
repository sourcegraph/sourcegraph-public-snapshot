package notebook

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/grafana/regexp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

// We must reimplement notebooks DB queries here because notebooks is an enterprise
// product. For some parts makes sense to re-implement some of this regardless since we
// want to specialize queries for search, but for e.g. permissions we might want to make
// the implementation OSS and use it from Enterprise. For now they are duplicated in
// store_copied.go
//
// TODO what is a good home for this?

type NotebooksSearchStore interface {
	SearchNotebooks(ctx context.Context, query string, searchBlocks bool) ([]*result.NotebookMatch, error)
}

type notebooksSearchStore struct {
	*basestore.Store
}

func Search(db dbutil.DB) NotebooksSearchStore {
	store := basestore.NewWithDB(db, sql.TxOptions{})
	return &notebooksSearchStore{store}
}

const searchNotebooksFmtStr = `
SELECT
	notebooks.id,
	notebooks.title,
	notebooks.blocks,
	NOT public as private, -- consistency with other match types
	users.username as namespace_user,
	orgs.name as namespace_org,
	(
		SELECT COUNT(*)
		FROM notebook_stars
		WHERE notebook_id = notebooks.id
	) as stars
FROM
	notebooks
	LEFT JOIN users on users.id = notebooks.namespace_user_id
	LEFT JOIN orgs on orgs.id = notebooks.namespace_org_id
WHERE
	(%s) -- permission conditions
	AND (%s) -- query conditions
ORDER BY
	stars DESC
LIMIT
	25
`

func scanMatch(scanner dbutil.Scanner) (*result.NotebookMatch, error) {
	n := &result.NotebookMatch{}
	var namespaceUser, namespaceOrg *string
	err := scanner.Scan(
		&n.ID,
		&n.Title,
		&n.Blocks,
		&n.Private,
		&namespaceUser,
		&namespaceOrg,
		&n.Stars,
	)
	if err != nil {
		return nil, err
	}
	if namespaceUser != nil {
		n.Namespace = *namespaceUser
	} else if namespaceOrg != nil {
		n.Namespace = *namespaceOrg
	}
	return n, nil
}

func scanNotebookMatches(rows *sql.Rows) ([]*result.NotebookMatch, error) {
	var notebooks []*result.NotebookMatch
	for rows.Next() {
		n, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		notebooks = append(notebooks, n)
	}
	return notebooks, nil
}

func makeQueryConds(query string, includeBlocks bool) *sqlf.Query {
	ilikeQuery := "%" + query + "%"

	conds := []*sqlf.Query{
		sqlf.Sprintf("CONCAT(users.username, orgs.name, notebooks.title) ILIKE %s",
			ilikeQuery),
	}
	if includeBlocks {
		// TODO this mirrors the GraphQL API implementation but does not allow us to
		// filter the blocks. we could potentially hack around this by doing the filtering
		// of blocks post-query by marshalling notebook blocks.
		conds = append(conds, sqlf.Sprintf("notebooks.blocks_tsvector @@ to_tsquery('english', %s)",
			toPostgresTextSearchQuery(query)))
	}

	return sqlf.Join(conds, "\n OR")
}

func (s *notebooksSearchStore) SearchNotebooks(ctx context.Context, query string, searchBlocks bool) ([]*result.NotebookMatch, error) {
	rows, err := s.Query(ctx,
		sqlf.Sprintf(
			searchNotebooksFmtStr,
			notebooksPermissionsCondition(ctx),
			makeQueryConds(query, true),
		),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	notebooks, err := scanNotebookMatches(rows)
	if err != nil {
		return nil, err
	}

	// TODO HACK post-filtering of blocks. Maybe this is okay because a notebook is
	// unlikely to have too many blocks, and we already know there is a match in here
	// because of `notebooks.blocks_tsvector`.
	if searchBlocks {
		searchRe, err := regexp.Compile("(?i).*(" + query + ").*")
		if err != nil {
			return nil, err
		}
		for _, n := range notebooks {
			var matchBlocks result.NotebookBlocks
			for _, block := range n.Blocks {
				// TODO this yolo marshals the entire block for search. If we had more
				// concrete types we can query specific fields, or emulate
				// `notebooks.blocks_tsvector` which takes all text-type fields.
				b, err := json.Marshal(block)
				if err != nil {
					continue
				}
				if searchRe.Match(b) {
					matchBlocks = append(matchBlocks, block)
				}
			}
			n.Blocks = matchBlocks
		}
	}

	return notebooks, nil
}
