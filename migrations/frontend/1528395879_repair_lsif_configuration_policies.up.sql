-- +++
-- parent: 1528395878
-- +++

BEGIN;

DROP TABLE lsif_configuration_policies;

CREATE TABLE lsif_configuration_policies (
    id SERIAL PRIMARY KEY,
    repository_id int,
    name text,
    type text NOT NULL,
    pattern text NOT NULL,
    retention_enabled boolean NOT NULL,
    retention_duration_hours int,
    retain_intermediate_commits boolean NOT NULL,
    indexing_enabled boolean NOT NULL,
    index_commit_max_age_hours int,
    index_intermediate_commits boolean NOT NULL
);

CREATE INDEX lsif_configuration_policies_repository_id ON lsif_configuration_policies(repository_id);

COMMENT ON COLUMN lsif_configuration_policies.repository_id IS 'The identifier of the repository to which this configuration policy applies. If absent, this policy is applied globally.';
COMMENT ON COLUMN lsif_configuration_policies.type IS 'The type of Git object (e.g., COMMIT, BRANCH, TAG).';
COMMENT ON COLUMN lsif_configuration_policies.pattern IS 'A pattern used to match` names of the associated Git object type.';
COMMENT ON COLUMN lsif_configuration_policies.retention_enabled IS 'Whether or not this configuration policy affects data retention rules.';
COMMENT ON COLUMN lsif_configuration_policies.retention_duration_hours IS 'The max age of data retained by this configuration policy. If null, the age is unbounded.';
COMMENT ON COLUMN lsif_configuration_policies.retain_intermediate_commits IS 'If the matching Git object is a branch, setting this value to true will also retain all data used to resolve queries for any commit on the matching branches. Setting this value to false will only consider the tip of the branch.';
COMMENT ON COLUMN lsif_configuration_policies.indexing_enabled IS 'Whether or not this configuration policy affects auto-indexing schedules.';
COMMENT ON COLUMN lsif_configuration_policies.index_commit_max_age_hours IS 'The max age of commits indexed by this configuration policy. If null, the age is unbounded.';
COMMENT ON COLUMN lsif_configuration_policies.index_intermediate_commits IS 'If the matching Git object is a branch, setting this value to true will also index all commits on the matching branches. Setting this value to false will only consider the tip of the branch.';

COMMIT;
