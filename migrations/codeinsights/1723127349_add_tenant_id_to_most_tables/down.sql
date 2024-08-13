-- Temporary function to deduplicate the logic required for each table:
CREATE OR REPLACE FUNCTION migrate_add_tenant_id_codeinsights(table_name text)
RETURNS void AS $$
BEGIN
    EXECUTE format('ALTER TABLE %I DROP COLUMN IF EXISTS tenant_id;', table_name);
END;
$$ LANGUAGE plpgsql;

SELECT migrate_add_tenant_id_codeinsights('archived_insight_series_recording_times'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('archived_series_points'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('dashboard'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('dashboard_grants'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('dashboard_insight_view'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('insight_series'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('insight_series_backfill'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('insight_series_incomplete_points'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('insight_series_recording_times'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('insight_view'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('insight_view_grants'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('insight_view_series'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('insights_background_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('insights_data_retention_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('metadata'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('repo_iterator'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('repo_iterator_errors'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('repo_names'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('series_points'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_codeinsights('series_points_snapshots'); COMMIT AND CHAIN;

DROP FUNCTION migrate_add_tenant_id_codeinsights(text);
