ALTER TABLE lsif_configuration_policies DROP COLUMN IF EXISTS lockfile_indexing_enabled;
ALTER TABLE lsif_configuration_policies ADD COLUMN IF NOT EXISTS embeddings_enabled BOOLEAN NOT NULL DEFAULT false;

CREATE OR REPLACE VIEW codeintel_configuration_policies AS
SELECT
    id,
    repository_id,
    name,
    type,
    pattern,
    retention_enabled,
    retention_duration_hours,
    retain_intermediate_commits,
    indexing_enabled,
    index_commit_max_age_hours,
    index_intermediate_commits,
    protected,
    repository_patterns,
    last_resolved_at,
    embeddings_enabled
FROM lsif_configuration_policies;

CREATE OR REPLACE VIEW codeintel_configuration_policies_repository_pattern_lookup AS
SELECT
    policy_id,
    repo_id
FROM lsif_configuration_policies_repository_pattern_lookup;
