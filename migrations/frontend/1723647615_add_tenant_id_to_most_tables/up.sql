-- This migration adds the tenant_id column in a way which doesn't require
-- updating every row. The value is null and an out of band migration will set
-- it to the default. A later migration will enforce tenant_id to be set.
--
-- We COMMIT AND CHAIN after each table is altered to prevent a single
-- transaction over all the alters. A single transaction would lead to a
-- deadlock with concurrent application queries.

-- Temporary function to deduplicate the logic required for each table:
CREATE OR REPLACE FUNCTION migrate_add_tenant_id_frontend(table_name text)
RETURNS void AS $$
BEGIN
    -- ALTER TABLE with a foreign key constraint will _always_ add the
    -- constraint, which means we always require a table lock even if this
    -- migration has run. So we check if the column exists first.
    IF NOT EXISTS (SELECT true
        FROM   pg_attribute
        WHERE  attrelid = quote_ident(table_name)::regclass
        AND    attname = 'tenant_id'
        AND    NOT attisdropped
    ) THEN
        EXECUTE format('ALTER TABLE %I ADD COLUMN IF NOT EXISTS tenant_id integer REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE CASCADE;', table_name);
    END IF;
END;
$$ LANGUAGE plpgsql;

SELECT migrate_add_tenant_id_frontend('access_requests'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('access_tokens'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('aggregated_user_statistics'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('assigned_owners'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('assigned_teams'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('batch_changes'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('batch_changes_site_credentials'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('batch_spec_execution_cache_entries'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('batch_spec_resolution_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('batch_spec_workspace_execution_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('batch_spec_workspace_execution_last_dequeues'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('batch_spec_workspace_files'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('batch_spec_workspaces'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('batch_specs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('cached_available_indexers'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('changeset_events'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('changeset_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('changeset_specs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('changesets'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('cm_action_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('cm_emails'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('cm_last_searched'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('cm_monitors'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('cm_queries'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('cm_recipients'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('cm_slack_webhooks'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('cm_trigger_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('cm_webhooks'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('code_hosts'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_autoindex_queue'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_autoindexing_exceptions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_commit_dates'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_inference_scripts'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_initial_path_ranks'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_initial_path_ranks_processed'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_langugage_support_requests'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_path_ranks'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_ranking_definitions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_ranking_exports'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_ranking_graph_keys'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_ranking_path_counts_inputs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_ranking_progress'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_ranking_references'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeintel_ranking_references_processed'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeowners'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeowners_individual_stats'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('codeowners_owners'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('commit_authors'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('configuration_policies_audit_logs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('context_detection_embedding_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('discussion_comments'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('discussion_mail_reply_tokens'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('discussion_threads'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('discussion_threads_target_repo'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('event_logs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('event_logs_export_allowlist'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('event_logs_scrape_state'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('event_logs_scrape_state_own'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('executor_heartbeats'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('executor_job_tokens'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('executor_secret_access_logs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('executor_secrets'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('exhaustive_search_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('exhaustive_search_repo_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('exhaustive_search_repo_revision_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('explicit_permissions_bitbucket_projects_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('external_service_repos'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('external_service_sync_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('external_services'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('feature_flag_overrides'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('feature_flags'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('github_app_installs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('github_apps'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('gitserver_relocator_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('gitserver_repos'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('gitserver_repos_statistics'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('gitserver_repos_sync_output'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('global_state'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('insights_query_runner_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('insights_query_runner_jobs_dependencies'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('insights_settings_migration_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('names'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('namespace_permissions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('notebook_stars'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('notebooks'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('org_invitations'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('org_members'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('org_stats'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('orgs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('orgs_open_beta_stats'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('out_of_band_migrations'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('out_of_band_migrations_errors'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('outbound_webhook_event_types'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('outbound_webhook_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('outbound_webhook_logs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('outbound_webhooks'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('own_aggregate_recent_contribution'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('own_aggregate_recent_view'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('own_background_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('own_signal_configurations'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('own_signal_recent_contribution'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('ownership_path_stats'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('package_repo_filters'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('package_repo_versions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('permission_sync_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('permissions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('phabricator_repos'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('product_licenses'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('product_subscriptions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('prompts'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('query_runner_state'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('redis_key_value'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('registry_extension_releases'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('registry_extensions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('repo'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('repo_commits_changelists'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('repo_embedding_job_stats'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('repo_embedding_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('repo_kvps'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('repo_paths'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('repo_pending_permissions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('repo_permissions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('repo_statistics'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('role_permissions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('roles'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('saved_searches'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('search_context_default'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('search_context_repos'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('search_context_stars'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('search_contexts'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('security_event_logs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('settings'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('sub_repo_permissions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('survey_responses'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('syntactic_scip_indexing_jobs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('syntactic_scip_last_index_scan'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('team_members'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('teams'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('telemetry_events_export_queue'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('temporary_settings'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('user_credentials'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('user_emails'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('user_external_accounts'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('user_onboarding_tour'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('user_pending_permissions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('user_permissions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('user_public_repos'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('user_repo_permissions'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('user_roles'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('users'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('vulnerabilities'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('vulnerability_affected_packages'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('vulnerability_affected_symbols'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('vulnerability_matches'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('webhook_logs'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('webhooks'); COMMIT AND CHAIN;
SELECT migrate_add_tenant_id_frontend('zoekt_repos'); COMMIT AND CHAIN;

-- Explicitly excluded tables
-- critical_and_site_config :: for instance not tenant
-- migration_logs :: about DB
-- tenants :: it is the foreign table
-- versions :: about the instance not the tenant
-- lsif_* :: many slow queries against them cause what looks like a deadlock

DROP FUNCTION migrate_add_tenant_id_frontend(text);
