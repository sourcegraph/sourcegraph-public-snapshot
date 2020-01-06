import { MAX_TRAVERSAL_LIMIT } from '../constants'

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
        -- Limit the visibility to the maximum traversal depth and approximate
        -- each commit's depth by its row number.
        limited_lineage AS (
            SELECT a.*, row_number() OVER() as n from lineage a LIMIT ${limit}
        ),
        -- Correlate commits to dumps and filter out commits without LSIF data
        lineage_with_dumps AS (
            SELECT a.*, d.root, d.id as dump_id FROM limited_lineage a
            JOIN lsif_dumps d ON d.repository = a.repository AND d."commit" = a."commit"
        ),
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
