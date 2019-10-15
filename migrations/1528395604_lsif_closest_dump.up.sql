BEGIN;

--
CREATE OR REPLACE FUNCTION closest_dump(repository text, "commit" text, traversal_limit integer) RETURNS SETOF lsif_dumps AS $$
    DECLARE
        r record;             -- lineage rows
        i float4 := 0;        -- traversal counter
        d lsif_dumps%ROWTYPE; -- lsif dump row (returned)
    BEGIN
        FOR r IN
            WITH RECURSIVE lineage(id, repository, "commit", parent_commit, direction) AS (
                -- seed result set with the target repository and commit marked by traversal direction
                SELECT l.* FROM (
                    SELECT c.*, 'A' FROM lsif_commits c WHERE c.repository = $1 AND c."commit" = $2
                    UNION
                    SELECT c.*, 'D' FROM lsif_commits c WHERE c.repository = $1 AND c."commit" = $2
                ) l
                UNION
                -- get the next commit in the ancestor or descendant direction
                SELECT * FROM (
                    WITH l_inner AS (SELECT * FROM lineage)
                    SELECT c.*, l.direction FROM l_inner l
                        JOIN lsif_commits c
                        ON l.direction = 'A' and c.repository = l.repository AND c.parent_commit = l."commit"
                    UNION
                    SELECT c.*, l.direction FROM l_inner l
                        JOIN lsif_commits c
                        ON l.direction = 'D' AND c.repository = l.repository AND c."commit" = l.parent_commit
                ) subquery
            )

            SELECT * FROM lineage l
        LOOP
            i := i + 1;

            IF i > $3 THEN
                RETURN;
            END IF;

            SELECT t.* INTO d FROM lsif_dumps t WHERE t.repository = r.repository AND t.commit = r.commit;
            IF d.id IS NOT NULL THEN
                RETURN NEXT d;
                RETURN;
            END IF;
        END LOOP;
    END;
$$ LANGUAGE plpgsql;

COMMIT;
