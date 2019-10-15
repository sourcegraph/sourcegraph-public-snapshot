BEGIN;

-- Find the dump associated with the commit closest to the given commit for which we have
-- LSIF data. This allows us to answer queries for commits sthat we do not have LSIF data
-- uploaded for, but do have LSIF data for a near ancestor or descendant. It is likely that
-- we can still use this data to provide (motly) precise code intelligence.
CREATE OR REPLACE FUNCTION closest_dump(repository text, "commit" text, traversal_limit integer) RETURNS SETOF lsif_dumps AS $$
    DECLARE
        r record;             -- lineage rows
        i float4 := 0;        -- traversal counter
        d lsif_dumps%ROWTYPE; -- lsif dump row (returned)
    BEGIN
        FOR r IN
            -- lineage is a recursively defined CTE that returns all ancestor an descendants
            -- of the given commit for the given repository. This happens to evaluate in
            -- Postgres as a lazy generator, which allows us to pull the "next" closest commit
            -- in either direction from the source commit a sneeded.
            WITH RECURSIVE lineage(id, repository, "commit", parent_commit, direction) AS (
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
                    -- get next ancestor
                    SELECT c.*, l.direction FROM l_inner l JOIN lsif_commits c ON l.direction = 'A' AND c.repository = l.repository AND c."commit" = l.parent_commit
                    UNION
                    -- get next descendant
                    SELECT c.*, l.direction FROM l_inner l JOIN lsif_commits c ON l.direction = 'D' and c.repository = l.repository AND c.parent_commit = l."commit"
                ) subquery
            )

            SELECT * FROM lineage l
        LOOP
            -- Keep track of how many rows we pull from lineage. If we hit the traversal
            -- limit, bail out now and return an empty row. Don't keep looking for a closest
            -- commit forever, as the farther we travel the more likely it is to be imprecise.
            i := i + 1;
            IF i > $3 THEN
                RETURN;
            END IF;

            -- Try to find the dump associated with this commit. If there is something
            -- return it and end the function. This will end up returning the first row
            -- we get from lineage with LSIF data which, as we pull rows back from
            -- lineage in order of distance from the source commit, it is necessarily
            -- the closest commit with LISF data.

            SELECT t.* INTO d FROM lsif_dumps t WHERE t.repository = r.repository AND t.commit = r.commit;
            IF d.id IS NOT NULL THEN
                RETURN NEXT d; -- return row
                RETURN;        -- exit function
            END IF;
        END LOOP;
    END;
$$ LANGUAGE plpgsql;

COMMIT;
