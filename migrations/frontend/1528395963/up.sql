CREATE TABLE IF NOT EXISTS insights_settings_migration_jobs
(
    id SERIAL NOT NULL,
    user_id int,
    org_id int,
    global boolean,
    settings_id int NOT NULL, -- non-constrained foreign key to settings object that should be migrated
    total_insights int NOT NULL DEFAULT 0,
    migrated_insights int NOT NULL DEFAULT 0,
    total_dashboards int NOT NULL DEFAULT 0,
    migrated_dashboards int NOT NULL DEFAULT 0,
    runs int NOT NULL DEFAULT 0,
    completed_at timestamp
);

TRUNCATE insights_settings_migration_jobs;

-- We go in this order (global, org, user) such that we migrate any higher level shared insights first. This way
-- we can just go in the order of id rather than have a secondary index.

-- global
INSERT INTO insights_settings_migration_jobs (settings_id, global)
SELECT id, TRUE
FROM settings
WHERE user_id IS NULL AND org_id IS NULL
ORDER BY id DESC
LIMIT 1;

-- org
INSERT INTO insights_settings_migration_jobs (settings_id, org_id)
SELECT DISTINCT ON (org_id) id, org_id
FROM settings
WHERE org_id IS NOT NULL
ORDER BY org_id, id DESC;

--  user
INSERT INTO insights_settings_migration_jobs (settings_id, user_id)
SELECT DISTINCT ON (user_id) id, user_id
FROM settings
WHERE user_id IS NOT NULL
ORDER BY user_id, id DESC;


INSERT INTO out_of_band_migrations(id, team, component, description, non_destructive,
                                   apply_reverse, is_enterprise, introduced_version_major, introduced_version_minor)
VALUES (14, 'code-insights', 'db.insights_settings_migration_jobs',
        'Migrating insight definitions from settings files to database tables as a last stage to use the GraphQL API.',
        TRUE, FALSE, TRUE, 3, 35)
ON CONFLICT DO NOTHING;
