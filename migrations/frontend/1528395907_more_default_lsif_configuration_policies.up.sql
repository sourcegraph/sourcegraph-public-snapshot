-- +++
-- parent: 1528395906
-- +++

BEGIN;

INSERT INTO lsif_configuration_policies
    (
        name,
        protected, pattern, type,
        retention_enabled, retain_intermediate_commits, retention_duration_hours,
        indexing_enabled, index_intermediate_commits, index_commit_max_age_hours
    )
VALUES
    (
        'Default commit retention policy',
        true, '*', 'GIT_TREE',
        true, true, 168, -- 1 week (168 hours) * 4 * 3
        false, false, 0
    );

COMMIT;
