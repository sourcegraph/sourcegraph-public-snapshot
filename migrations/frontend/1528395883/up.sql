-- +++
-- parent: 1528395882
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
        'Default tip-of-branch retention policy',
        true, '*', 'GIT_TREE',
        true, false, 2016, -- 3 months ~= 2016 hours = 1 week (168 hours) * 4 * 3
        false, false, 0
    ), (
        'Default tag retention policy',
        true, '*', 'GIT_TAG',
        true, false, 8064, -- 12 months ~= 8064 hours = 1 week (168 hours) * 4 * 12
        false, false, 0
    );

COMMIT;
