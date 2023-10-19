-- Perform migration here.
--
-- See /migrations/README.md. Highlights:
--  * Make migrations idempotent (use IF EXISTS)
--  * Make migrations backwards-compatible (old readers/writers must continue to work)
--  * If you are using CREATE INDEX CONCURRENTLY, then make sure that only one statement
--    is defined per file, and that each such statement is NOT wrapped in a transaction.
--    Each such migration must also declare "createIndexConcurrently: true" in their
--    associated metadata.yaml file.
--  * If you are modifying Postgres extensions, you must also declare "privileged: true"
--    in the associated metadata.yaml file.

ALTER TABLE IF EXISTS repo
ADD COLUMN IF NOT EXISTS topics text[] GENERATED ALWAYS AS (
    CASE
    WHEN repo.external_service_type = 'github' THEN
        jsonb_array_elements_text(jsonb_path_query_array(repo.metadata, '$.RepositoryTopics.Nodes[*].Topic.Name'))::text[]
    WHEN repo.external_service_type = 'gitlab' THEN
        jsonb_array_elements_text(repo.metadata->'topics')::text[]
    ELSE
        '{}'::text[]
    END
) STORED;
