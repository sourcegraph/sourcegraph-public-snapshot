-- +++
-- parent: 1528395918
-- +++

BEGIN;

DELETE FROM search_contexts WHERE deleted_at IS NOT NULL;

COMMENT ON COLUMN search_contexts.deleted_at IS 'This column is unused as of Sourcegraph 3.34. Do not refer to it anymore. It will be dropped in a future version.';

ALTER TABLE search_context_repos ADD CONSTRAINT search_context_repos_unique UNIQUE (repo_id, search_context_id, revision);

ALTER TABLE search_context_repos DROP CONSTRAINT IF EXISTS search_context_repos_search_context_id_repo_id_revision_unique;

COMMIT;
