BEGIN;

DROP VIEW IF EXISTS lsif_commits_with_lsif_data;

-- now there can be multiple LSIF dumps per repository@commit.
ALTER TABLE lsif_dumps DROP CONSTRAINT lsif_dumps_repository_commit_key;
ALTER TABLE lsif_dumps ADD COLUMN root TEXT NOT NULL DEFAULT '';
ALTER TABLE lsif_dumps ADD CONSTRAINT lsif_dumps_repository_commit_root UNIQUE (repository, commit, root);

-- Find the dump associated with the commit closest to the given commit for which we have
-- LSIF data. This allows us to answer queries for commits that we do not have LSIF data
-- uploaded for, but do have LSIF data for a near ancestor or descendant. It is likely that
-- we can still use this data to provide (mostly) precise code intelligence.
CREATE OR REPLACE FUNCTION closest_dump(repository text, "commit" text, path text, traversal_limit integer) RETURNS SETOF lsif_dumps AS $$
    DECLARE
        lineage_row record;            -- lineage rows
        i float4 := 0;                 -- traversal counter
        found_dump lsif_dumps%ROWTYPE; -- lsif dump row (returned)
    BEGIN
        FOR lineage_row IN
            -- lineage is a recursively defined CTE that returns all ancestor an descendants
            -- of the given commit for the given repository. This happens to evaluate in
            -- Postgres as a lazy generator, which allows us to pull the "next" closest commit
            -- in either direction from the source commit as needed.
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
            IF i > $4 THEN
                RETURN;
            END IF;

            -- Try to find the dump associated with this commit. If there is something
            -- return it and end the function. This will end up returning the first row
            -- we get from lineage with LSIF data which, as we pull rows back from
            -- lineage in order of distance from the source commit, it is necessarily
            -- the closest commit with LSIF data.

            SELECT dump.* INTO found_dump FROM lsif_dumps dump WHERE dump.repository = lineage_row.repository AND dump.commit = lineage_row.commit AND $3 LIKE (dump.root || '%');
            IF found_dump.id IS NOT NULL THEN
                RETURN NEXT found_dump; -- return row
                RETURN;                 -- exit function
            END IF;
        END LOOP;
    END;
$$ LANGUAGE plpgsql;

END;
