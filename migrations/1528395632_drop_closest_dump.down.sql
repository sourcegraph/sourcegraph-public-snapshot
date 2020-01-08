-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

-- Re-define function defined in 1528395604_lsif_multi_project.
CREATE FUNCTION closest_dump(repository text, "commit" text, path text, traversal_limit integer) RETURNS SETOF lsif_dumps AS $$
    DECLARE
        lineage_row record;
        i float4 := 0;
        found_dump lsif_dumps%ROWTYPE;
    BEGIN
        FOR lineage_row IN
            WITH RECURSIVE lineage(id, repository, "commit", parent_commit, direction) AS (
                SELECT l.* FROM (
                    SELECT c.*, 'A' FROM lsif_commits c WHERE c.repository = $1 AND c."commit" = $2
                    UNION
                    SELECT c.*, 'D' FROM lsif_commits c WHERE c.repository = $1 AND c."commit" = $2
                ) l

                UNION

                SELECT * FROM (
                    WITH l_inner AS (SELECT * FROM lineage)
                    SELECT c.*, l.direction FROM l_inner l JOIN lsif_commits c ON l.direction = 'A' AND c.repository = l.repository AND c."commit" = l.parent_commit
                    UNION
                    SELECT c.*, l.direction FROM l_inner l JOIN lsif_commits c ON l.direction = 'D' and c.repository = l.repository AND c.parent_commit = l."commit"
                ) subquery
            )

            SELECT * FROM lineage l
        LOOP
            i := i + 1;
            IF i > $4 THEN
                RETURN;
            END IF;

            SELECT dump.* INTO found_dump FROM lsif_dumps dump WHERE dump.repository = lineage_row.repository AND dump.commit = lineage_row.commit AND $3 LIKE (dump.root || '%');
            IF found_dump.id IS NOT NULL THEN
                RETURN NEXT found_dump;
                RETURN;
            END IF;
        END LOOP;
    END;
$$ LANGUAGE plpgsql;

COMMIT;
