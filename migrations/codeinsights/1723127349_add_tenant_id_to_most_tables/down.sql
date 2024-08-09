-- Temporary function to deduplicate the logic required for each table:
CREATE OR REPLACE FUNCTION migrate_table(table_name text)
RETURNS void AS $$
BEGIN
    EXECUTE format('ALTER TABLE %I DROP COLUMN IF EXISTS tenant_id;', table_name);
END;
$$ LANGUAGE plpgsql;

SELECT migrate_table('archived_insight_series_recording_times');
SELECT migrate_table('archived_series_points');
SELECT migrate_table('dashboard');
SELECT migrate_table('dashboard_grants');
SELECT migrate_table('dashboard_insight_view');
SELECT migrate_table('insight_series');
SELECT migrate_table('insight_series_backfill');
SELECT migrate_table('insight_series_incomplete_points');
SELECT migrate_table('insight_series_recording_times');
SELECT migrate_table('insight_view');
SELECT migrate_table('insight_view_grants');
SELECT migrate_table('insight_view_series');
SELECT migrate_table('insights_background_jobs');
SELECT migrate_table('insights_data_retention_jobs');
SELECT migrate_table('metadata');
SELECT migrate_table('repo_iterator');
SELECT migrate_table('repo_iterator_errors');
SELECT migrate_table('repo_names');
SELECT migrate_table('series_points');
SELECT migrate_table('series_points_snapshots');

DROP FUNCTION migrate_table(text);
