BEGIN;

ALTER TABLE search_context_repos ADD CONSTRAINT search_context_repos_search_context_id_repo_id_revision_unique UNIQUE (search_context_id, repo_id, revision);

ALTER TABLE search_context_repos DROP CONSTRAINT IF EXISTS search_context_repos_unique;

COMMIT;
