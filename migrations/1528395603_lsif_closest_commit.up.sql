-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

-- Retrieve a commit that has LSIF data for the given repository. This commit is chosen by
-- traversing the ancestors and descendant commits of the origin commit. The first commit
-- with LSIF data seen is returned. The nubmer of commits traversed will not exceed the
-- given traversal limit.

CREATE OR REPLACE FUNCTION lsif_closest_commit_with_data(repository text, "commit" text, traversal_limit integer) RETURNS text
AS $$
    WITH RECURSIVE lineage(repository, "commit", parent_commit, has_lsif_data, direction) AS (
        -- seed result set with the target repository and commit marked by traversal direction
        SELECT l.* FROM (
            SELECT c.*, 'A' FROM lsif_commits_with_lsif_data c WHERE c.repository = $1 AND c."commit" = $2
            UNION
            SELECT c.*, 'D' FROM lsif_commits_with_lsif_data c WHERE c.repository = $1 AND c."commit" = $2
        ) l

        UNION

        -- get the next commit in the ancestor or descendant direction
        SELECT * FROM (
            WITH l_inner AS (SELECT * FROM lineage)
            SELECT c.*, l.direction FROM l_inner l JOIN lsif_commits_with_lsif_data c ON l.direction = 'A' and c.repository = l.repository AND c.parent_commit = l."commit"
            UNION
            SELECT c.*, l.direction FROM l_inner l JOIN lsif_commits_with_lsif_data c ON l.direction = 'D' AND c.repository = l.repository AND c."commit" = l.parent_commit
        ) subquery
    )

    -- lineage is ordered by distance to the target commit by construction; get first commit with data
    SELECT l.commit FROM (SELECT * FROM lineage LIMIT $3) l WHERE l.has_lsif_data LIMIT 1;
$$ LANGUAGE SQL;

COMMIT;
