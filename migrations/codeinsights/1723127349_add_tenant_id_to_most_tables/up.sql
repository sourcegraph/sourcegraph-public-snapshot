-- This migration adds the tenant_id column in a way which doesn't require
-- updating every row. The value is null and an out of band migration will set
-- it to the default. A later migration will enforce tenant_id to be set.
--
-- We wrap each call with its own transaction. This is since mutliple
-- statements in a single call to postgres is run in its own transaction by
-- default. We want to run each alter without joining the table locks with
-- eachother to avoid deadlock.

-- Temporary function to deduplicate the logic required for each table:
CREATE OR REPLACE FUNCTION migrate_add_tenant_id_codeinsights(table_name text)
RETURNS void AS $$
BEGIN
    EXECUTE format('ALTER TABLE %I ADD COLUMN IF NOT EXISTS tenant_id integer REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE CASCADE;', table_name);
END;
$$ LANGUAGE plpgsql;

BEGIN; SELECT migrate_add_tenant_id_codeinsights('archived_insight_series_recording_times'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('archived_series_points'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('dashboard'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('dashboard_grants'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('dashboard_insight_view'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('insight_series'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('insight_series_backfill'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('insight_series_incomplete_points'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('insight_series_recording_times'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('insight_view'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('insight_view_grants'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('insight_view_series'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('insights_background_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('insights_data_retention_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('metadata'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('repo_iterator'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('repo_iterator_errors'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('repo_names'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('series_points'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_codeinsights('series_points_snapshots'); COMMIT;

-- Explicitly excluded tables
-- migration_logs :: about DB

DROP FUNCTION migrate_add_tenant_id_codeinsights(text);
