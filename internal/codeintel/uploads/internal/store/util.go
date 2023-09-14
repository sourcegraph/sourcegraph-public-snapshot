package store

import (
	"database/sql"
	"sort"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

const uploadRankQueryFragment = `
SELECT
	r.id,
	ROW_NUMBER() OVER (ORDER BY
		-- Note: this should be kept in-sync with the order given to workerutil
		r.associated_index_id IS NULL DESC,
		COALESCE(r.process_after, r.uploaded_at),
		r.id
	) as rank
FROM lsif_uploads_with_repository_name r
WHERE r.state = 'queued'
`

const visibleAtTipSubselectQuery = `
SELECT 1
FROM lsif_uploads_visible_at_tip uvt
WHERE
	uvt.repository_id = u.repository_id AND
	uvt.upload_id = u.id AND
	uvt.is_default_branch
`

// packageRankingQueryFragment uses `lsif_uploads u` JOIN `lsif_packages p` to return a rank
// for each row grouped by package and source code location and ordered by the associated Git
// commit date.
const packageRankingQueryFragment = `
rank() OVER (
	PARTITION BY
		-- Group providers of the same package together
		p.scheme, p.manager, p.name, p.version,
		-- Defined by the same directory within a repository
		u.repository_id, u.indexer, u.root
	ORDER BY
		-- Rank each grouped upload by the associated commit date
		(SELECT cd.committed_at FROM codeintel_commit_dates cd WHERE cd.repository_id = u.repository_id AND cd.commit_bytea = decode(u.commit, 'hex')) NULLS LAST,
		-- Break ties via the unique identifier
		u.id
)
`

const indexAssociatedUploadIDQueryFragment = `
(
	SELECT MAX(id) FROM lsif_uploads WHERE associated_index_id = u.id
) AS associated_upload_id
`

const indexRankQueryFragment = `
SELECT
	r.id,
	ROW_NUMBER() OVER (ORDER BY COALESCE(r.process_after, r.queued_at), r.id) as rank
FROM lsif_indexes_with_repository_name r
WHERE r.state = 'queued'
`

// makeVisibleUploadsQuery returns a SQL query returning the set of identifiers of uploads
// visible from the given commit. This is done by removing the "shadowed" values created
// by looking at a commit and it's ancestors visible commits.
func makeVisibleUploadsQuery(repositoryID int, commit string) *sqlf.Query {
	return sqlf.Sprintf(
		visibleUploadsQuery,
		repositoryID, dbutil.CommitBytea(commit),
		repositoryID, dbutil.CommitBytea(commit),
	)
}

const visibleUploadsQuery = `
SELECT
	t.upload_id
FROM (
	SELECT
		t.*,
		row_number() OVER (PARTITION BY root, indexer ORDER BY distance) AS r
	FROM (
		-- Select the set of uploads visible from the given commit. This is done by looking
		-- at each commit's row in the lsif_nearest_uploads table, and the (adjusted) set of
		-- uploads from each commit's nearest ancestor according to the data compressed in
		-- the links table.
		--
		-- NB: A commit should be present in at most one of these tables.
		SELECT
			nu.repository_id,
			upload_id::integer,
			nu.commit_bytea,
			u_distance::text::integer as distance
		FROM lsif_nearest_uploads nu
		CROSS JOIN jsonb_each(nu.uploads) as u(upload_id, u_distance)
		WHERE nu.repository_id = %s AND nu.commit_bytea = %s
		UNION (
			SELECT
				nu.repository_id,
				upload_id::integer,
				ul.commit_bytea,
				u_distance::text::integer + ul.distance as distance
			FROM lsif_nearest_uploads_links ul
			JOIN lsif_nearest_uploads nu ON nu.repository_id = ul.repository_id AND nu.commit_bytea = ul.ancestor_commit_bytea
			CROSS JOIN jsonb_each(nu.uploads) as u(upload_id, u_distance)
			WHERE nu.repository_id = %s AND ul.commit_bytea = %s
		)
	) t
	JOIN lsif_uploads u ON u.id = upload_id
) t
-- Remove ranks > 1, as they are shadowed by another upload in the same output set
WHERE t.r <= 1
`

func scanCountsWithTotalCount(rows *sql.Rows, queryErr error) (totalCount int, _ map[int]int, err error) {
	if queryErr != nil {
		return 0, nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	visibilities := map[int]int{}
	for rows.Next() {
		var id int
		var count int
		if err := rows.Scan(&totalCount, &id, &count); err != nil {
			return 0, nil, err
		}

		visibilities[id] = count
	}

	return totalCount, visibilities, nil
}

// scanSourcedCommits scans triples of repository ids/repository names/commits from the
// return value of `*Store.query`. The output of this function is ordered by repository
// identifier, then by commit.
func scanSourcedCommits(rows *sql.Rows, queryErr error) (_ []SourcedCommits, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	sourcedCommitsMap := map[int]SourcedCommits{}
	for rows.Next() {
		var repositoryID int
		var repositoryName string
		var commit string
		if err := rows.Scan(&repositoryID, &repositoryName, &commit); err != nil {
			return nil, err
		}

		sourcedCommitsMap[repositoryID] = SourcedCommits{
			RepositoryID:   repositoryID,
			RepositoryName: repositoryName,
			Commits:        append(sourcedCommitsMap[repositoryID].Commits, commit),
		}
	}

	flattened := make([]SourcedCommits, 0, len(sourcedCommitsMap))
	for _, sourcedCommits := range sourcedCommitsMap {
		sort.Strings(sourcedCommits.Commits)
		flattened = append(flattened, sourcedCommits)
	}

	sort.Slice(flattened, func(i, j int) bool {
		return flattened[i].RepositoryID < flattened[j].RepositoryID
	})
	return flattened, nil
}
