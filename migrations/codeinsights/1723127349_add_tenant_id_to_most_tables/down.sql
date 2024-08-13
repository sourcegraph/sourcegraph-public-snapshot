-- Temporary function to deduplicate the logic required for each table:
CREATE OR REPLACE FUNCTION migrate_add_tenant_id_codeinsights(table_name text)
RETURNS void AS $$
BEGIN
    EXECUTE format('ALTER TABLE %I DROP COLUMN IF EXISTS tenant_id;', table_name);
END;
$$ LANGUAGE plpgsql;

SELECT migrate_add_tenant_id_codeinsights('archived_insight_series_recording_times');
SELECT migrate_add_tenant_id_codeinsights('archived_series_points');
SELECT migrate_add_tenant_id_codeinsights('dashboard');
SELECT migrate_add_tenant_id_codeinsights('dashboard_grants');
SELECT migrate_add_tenant_id_codeinsights('dashboard_insight_view');
SELECT migrate_add_tenant_id_codeinsights('insight_series');
SELECT migrate_add_tenant_id_codeinsights('insight_series_backfill');
SELECT migrate_add_tenant_id_codeinsights('insight_series_incomplete_points');
SELECT migrate_add_tenant_id_codeinsights('insight_series_recording_times');
SELECT migrate_add_tenant_id_codeinsights('insight_view');
SELECT migrate_add_tenant_id_codeinsights('insight_view_grants');
SELECT migrate_add_tenant_id_codeinsights('insight_view_series');
SELECT migrate_add_tenant_id_codeinsights('insights_background_jobs');
SELECT migrate_add_tenant_id_codeinsights('insights_data_retention_jobs');
SELECT migrate_add_tenant_id_codeinsights('metadata');
SELECT migrate_add_tenant_id_codeinsights('repo_iterator');
SELECT migrate_add_tenant_id_codeinsights('repo_iterator_errors');
SELECT migrate_add_tenant_id_codeinsights('repo_names');
SELECT migrate_add_tenant_id_codeinsights('series_points');
SELECT migrate_add_tenant_id_codeinsights('series_points_snapshots');

DROP FUNCTION migrate_add_tenant_id_codeinsights(text);
