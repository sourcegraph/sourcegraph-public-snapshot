import { MAX_TRAVERSAL_LIMIT } from '../constants'

/**
 * Return a recursive CTE `lineage` that returns ancestors of the commit for the given
 * repository. This assumes that the repo name is $1 and the commit is $2.
 */
export function ancestorLineage(): string {
    return `
        RECURSIVE lineage(id, repository, "commit", parent) AS (
            SELECT c.* FROM lsif_commits c WHERE c.repository = $1 AND c."commit" = $2
            UNION
            SELECT c.* FROM lineage a JOIN lsif_commits c ON a.repository = c.repository AND a.parent = c."commit"
        )
    `
}

/**
 * Return a recursive CTE `lineage` that returns ancestors and descendants of the commit for
 * the given repository. This assumes that the repo name is $1 and the commit is $2. This
 * happens to evaluate in Postgres as a lazy generator, which allows us to pull the "next"
 * closest commit in either direction from the source commit as needed.
 */
export function bidirectionalLineage(): string {
    return `
        RECURSIVE lineage(id, repository, "commit", parent_commit, direction) AS (
            SELECT l.* FROM (
                -- seed recursive set with commit looking in ancestor direction
                SELECT c.*, 'A' FROM lsif_commits c WHERE c.repository = $1 AND c."commit" = $2
                UNION
                -- seed recursive set with commit looking in descendant direction
                SELECT c.*, 'D' FROM lsif_commits c WHERE c.repository = $1 AND c."commit" = $2
            ) l

            UNION

            SELECT * FROM (
                WITH l_inner AS (SELECT * FROM lineage)
                -- get next ancestors (multiple parents for merge commits)
                SELECT c.*, 'A' FROM l_inner l JOIN lsif_commits c ON l.direction = 'A' AND c.repository = l.repository AND c."commit" = l.parent_commit
                UNION
                -- get next descendants
                SELECT c.*, 'D' FROM l_inner l JOIN lsif_commits c ON l.direction = 'D' and c.repository = l.repository AND c.parent_commit = l."commit"
            ) subquery
        )
    `
}

/**
 * Return a set of CTE definitions assuming the definition of a previous CTE named `lineage`.
 * This creates the CTE `lineage_with_dumps`, which gathers the set of LSIF dump identifiers
 * whose commit occurs in `lineage` (within the given traversal limit).
 *
 * @param limit The maximum number of dumps that can be extracted from `lineage`.
 */
export function lineageWithDumps(limit: number = MAX_TRAVERSAL_LIMIT): string {
    return `
        -- Limit the visibility to the maximum traversal depth and approximate
        -- each commit's depth by its row number.
        limited_lineage AS (
            SELECT a.*, row_number() OVER() as n from lineage a LIMIT ${limit}
        ),
        -- Correlate commits to dumps and filter out commits without LSIF data
        lineage_with_dumps AS (
            SELECT a.*, d.root, d.id as dump_id FROM limited_lineage a
            JOIN lsif_dumps d ON d.repository = a.repository AND d."commit" = a."commit"
        )
    `
}

/**
 * Return a set of CTE definitions assuming the definition of a previous CTE named `lineage`.
 * This creates the CTE `visible_ids`, which gathers the set of LSIF dump identifiers whose
 * commit occurs in `lineage` (within the given traversal limit) and whose root does not
 * overlap another visible dump.
 *
 * @param limit The maximum number of dumps that can be extracted from `lineage`.
 */
export function visibleDumps(limit: number = MAX_TRAVERSAL_LIMIT): string {
    return `
        ${lineageWithDumps(limit)},
        visible_ids AS (
            -- Remove dumps where there exists another visible dump of smaller depth with an overlapping root.
            -- Such dumps would not be returned with a closest commit query so we don't want to return results
            -- for them in global find-reference queries either.
            SELECT DISTINCT t1.dump_id as id FROM lineage_with_dumps t1 WHERE NOT EXISTS (
                SELECT 1 FROM lineage_with_dumps t2
                WHERE t2.n < t1.n AND (
                    t2.root LIKE (t1.root || '%') OR
                    t1.root LIKE (t2.root || '%')
                )
            )
        )
    `
}
