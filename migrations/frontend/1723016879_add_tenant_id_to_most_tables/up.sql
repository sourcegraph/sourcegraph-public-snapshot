-- This migration adds the tenant_id column in a way which doesn't require
-- updating every row. The value is null and an out of band migration will set
-- it to the default. A later migration will enforce tenant_id to be set.
--
-- We wrap each call with its own transaction. This is since mutliple
-- statements in a single call to postgres is run in its own transaction by
-- default. We want to run each alter without joining the table locks with
-- eachother to avoid deadlock.

-- Temporary function to deduplicate the logic required for each table:
CREATE OR REPLACE FUNCTION migrate_add_tenant_id_frontend(table_name text)
RETURNS void AS $$
BEGIN
    EXECUTE format('ALTER TABLE %I ADD COLUMN IF NOT EXISTS tenant_id integer REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE CASCADE;', table_name);
END;
$$ LANGUAGE plpgsql;

BEGIN; SELECT migrate_add_tenant_id_frontend('access_requests'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('access_tokens'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('aggregated_user_statistics'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('assigned_owners'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('assigned_teams'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('batch_changes'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('batch_changes_site_credentials'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('batch_spec_execution_cache_entries'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('batch_spec_resolution_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('batch_spec_workspace_execution_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('batch_spec_workspace_execution_last_dequeues'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('batch_spec_workspace_files'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('batch_spec_workspaces'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('batch_specs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('cached_available_indexers'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('changeset_events'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('changeset_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('changeset_specs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('changesets'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('cm_action_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('cm_emails'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('cm_last_searched'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('cm_monitors'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('cm_queries'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('cm_recipients'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('cm_slack_webhooks'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('cm_trigger_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('cm_webhooks'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('code_hosts'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_autoindex_queue'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_autoindexing_exceptions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_commit_dates'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_inference_scripts'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_initial_path_ranks'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_initial_path_ranks_processed'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_langugage_support_requests'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_path_ranks'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_ranking_definitions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_ranking_exports'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_ranking_graph_keys'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_ranking_path_counts_inputs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_ranking_progress'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_ranking_references'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeintel_ranking_references_processed'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeowners'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeowners_individual_stats'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('codeowners_owners'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('commit_authors'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('configuration_policies_audit_logs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('context_detection_embedding_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('discussion_comments'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('discussion_mail_reply_tokens'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('discussion_threads'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('discussion_threads_target_repo'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('event_logs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('event_logs_export_allowlist'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('event_logs_scrape_state'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('event_logs_scrape_state_own'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('executor_heartbeats'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('executor_job_tokens'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('executor_secret_access_logs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('executor_secrets'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('exhaustive_search_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('exhaustive_search_repo_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('exhaustive_search_repo_revision_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('explicit_permissions_bitbucket_projects_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('external_service_repos'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('external_service_sync_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('external_services'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('feature_flag_overrides'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('feature_flags'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('github_app_installs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('github_apps'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('gitserver_relocator_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('gitserver_repos'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('gitserver_repos_statistics'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('gitserver_repos_sync_output'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('global_state'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('insights_query_runner_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('insights_query_runner_jobs_dependencies'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('insights_settings_migration_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_configuration_policies'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_configuration_policies_repository_pattern_lookup'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_dependency_indexing_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_dependency_repos'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_dependency_syncing_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_dirty_repositories'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_index_configuration'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_indexes'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_last_index_scan'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_last_retention_scan'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_nearest_uploads'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_nearest_uploads_links'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_packages'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_references'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_retention_configuration'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_uploads'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_uploads_audit_logs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_uploads_reference_counts'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_uploads_visible_at_tip'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('lsif_uploads_vulnerability_scan'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('names'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('namespace_permissions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('notebook_stars'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('notebooks'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('org_invitations'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('org_members'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('org_stats'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('orgs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('orgs_open_beta_stats'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('out_of_band_migrations'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('out_of_band_migrations_errors'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('outbound_webhook_event_types'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('outbound_webhook_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('outbound_webhook_logs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('outbound_webhooks'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('own_aggregate_recent_contribution'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('own_aggregate_recent_view'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('own_background_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('own_signal_configurations'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('own_signal_recent_contribution'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('ownership_path_stats'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('package_repo_filters'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('package_repo_versions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('permission_sync_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('permissions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('phabricator_repos'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('product_licenses'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('product_subscriptions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('prompts'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('query_runner_state'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('redis_key_value'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('registry_extension_releases'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('registry_extensions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('repo'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('repo_commits_changelists'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('repo_embedding_job_stats'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('repo_embedding_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('repo_kvps'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('repo_paths'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('repo_pending_permissions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('repo_permissions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('repo_statistics'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('role_permissions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('roles'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('saved_searches'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('search_context_default'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('search_context_repos'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('search_context_stars'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('search_contexts'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('security_event_logs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('settings'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('sub_repo_permissions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('survey_responses'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('syntactic_scip_indexing_jobs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('syntactic_scip_last_index_scan'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('team_members'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('teams'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('telemetry_events_export_queue'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('temporary_settings'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('user_credentials'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('user_emails'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('user_external_accounts'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('user_onboarding_tour'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('user_pending_permissions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('user_permissions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('user_public_repos'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('user_repo_permissions'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('user_roles'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('users'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('vulnerabilities'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('vulnerability_affected_packages'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('vulnerability_affected_symbols'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('vulnerability_matches'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('webhook_logs'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('webhooks'); COMMIT;
BEGIN; SELECT migrate_add_tenant_id_frontend('zoekt_repos'); COMMIT;

-- Explicitly excluded tables
-- critical_and_site_config :: for instance not tenant
-- migration_logs :: about DB
-- tenants :: it is the foreign table
-- versions :: about the instance not the tenant

DROP FUNCTION migrate_add_tenant_id_frontend(text);
