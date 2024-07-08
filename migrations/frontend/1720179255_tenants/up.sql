CREATE TABLE IF NOT EXISTS tenants (
    id SERIAL PRIMARY KEY,
    name text NOT NULL,
    slug text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

INSERT INTO tenants (id, name, slug, created_at, updated_at) VALUES (1, 'default', 'default', NOW(), NOW()) ON CONFLICT (id) DO NOTHING;
SET app.current_tenant = 1;

-- Temporary function to deduplicate the above queries for each table:
CREATE OR REPLACE FUNCTION migrate_table(table_name text)
RETURNS void AS $$
BEGIN
    -- todo: non nullable column?
    EXECUTE format('
        ALTER TABLE %I ADD COLUMN IF NOT EXISTS tenant_id integer DEFAULT current_setting(''app.current_tenant'')::integer REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE CASCADE;
        -- TODO REMOVE
        UPDATE %I SET tenant_id = 1;
        ALTER TABLE %I ALTER COLUMN tenant_id SET NOT NULL;
        ALTER TABLE %I ENABLE ROW LEVEL SECURITY;
        DROP POLICY IF EXISTS %I_isolation_policy ON %I;
        CREATE POLICY %I_isolation_policy ON %I USING (tenant_id = current_setting(''app.current_tenant'')::integer);
        ALTER TABLE %I FORCE ROW LEVEL SECURITY;',
        table_name, table_name, table_name, table_name, table_name, table_name, table_name, table_name, table_name
    );
END;
$$ LANGUAGE plpgsql;

-- todo, those might all need to run in separate transactions, as pending triggers can fail here.
SELECT migrate_table('access_requests');
SELECT migrate_table('access_tokens');
SELECT migrate_table('aggregated_user_statistics');
SELECT migrate_table('assigned_owners');
SELECT migrate_table('assigned_teams');
SELECT migrate_table('batch_changes');
SELECT migrate_table('batch_changes_site_credentials');
SELECT migrate_table('batch_spec_execution_cache_entries');
SELECT migrate_table('batch_spec_resolution_jobs');
SELECT migrate_table('batch_spec_workspace_execution_jobs');
SELECT migrate_table('batch_spec_workspace_execution_last_dequeues');
SELECT migrate_table('batch_spec_workspace_files');
SELECT migrate_table('batch_spec_workspaces');
SELECT migrate_table('batch_specs');
SELECT migrate_table('cached_available_indexers');
SELECT migrate_table('changeset_events');
SELECT migrate_table('changeset_jobs');
SELECT migrate_table('changeset_specs');
SELECT migrate_table('changesets');
SELECT migrate_table('cm_action_jobs');
SELECT migrate_table('cm_emails');
SELECT migrate_table('cm_last_searched');
SELECT migrate_table('cm_monitors');
SELECT migrate_table('cm_queries');
SELECT migrate_table('cm_recipients');
SELECT migrate_table('cm_slack_webhooks');
SELECT migrate_table('cm_trigger_jobs');
SELECT migrate_table('cm_webhooks');
SELECT migrate_table('code_hosts');
SELECT migrate_table('codeintel_autoindex_queue');
SELECT migrate_table('codeintel_autoindexing_exceptions');
SELECT migrate_table('codeintel_commit_dates');
SELECT migrate_table('codeintel_inference_scripts');
SELECT migrate_table('codeintel_initial_path_ranks');
SELECT migrate_table('codeintel_initial_path_ranks_processed');
SELECT migrate_table('codeintel_langugage_support_requests');
SELECT migrate_table('codeintel_path_ranks');
SELECT migrate_table('codeintel_ranking_definitions');
SELECT migrate_table('codeintel_ranking_exports');
SELECT migrate_table('codeintel_ranking_graph_keys');
SELECT migrate_table('codeintel_ranking_path_counts_inputs');
SELECT migrate_table('codeintel_ranking_progress');
SELECT migrate_table('codeintel_ranking_references');
SELECT migrate_table('codeintel_ranking_references_processed');
SELECT migrate_table('codeowners');
SELECT migrate_table('codeowners_individual_stats');
SELECT migrate_table('codeowners_owners');
SELECT migrate_table('commit_authors');
SELECT migrate_table('configuration_policies_audit_logs');
SELECT migrate_table('context_detection_embedding_jobs');
SELECT migrate_table('critical_and_site_config');
SELECT migrate_table('discussion_comments');
SELECT migrate_table('discussion_mail_reply_tokens');
SELECT migrate_table('discussion_threads');
SELECT migrate_table('discussion_threads_target_repo');
SELECT migrate_table('event_logs');
SELECT migrate_table('event_logs_export_allowlist');
SELECT migrate_table('event_logs_scrape_state');
SELECT migrate_table('event_logs_scrape_state_own');
SELECT migrate_table('executor_heartbeats');
SELECT migrate_table('executor_job_tokens');
SELECT migrate_table('executor_secret_access_logs');
SELECT migrate_table('executor_secrets');
SELECT migrate_table('exhaustive_search_jobs');
SELECT migrate_table('exhaustive_search_repo_jobs');
SELECT migrate_table('exhaustive_search_repo_revision_jobs');
SELECT migrate_table('explicit_permissions_bitbucket_projects_jobs');
SELECT migrate_table('external_service_repos');
SELECT migrate_table('external_service_sync_jobs');
SELECT migrate_table('external_services');
SELECT migrate_table('feature_flag_overrides');
SELECT migrate_table('feature_flags');
SELECT migrate_table('github_app_installs');
SELECT migrate_table('github_apps');
SELECT migrate_table('gitserver_relocator_jobs');
SELECT migrate_table('gitserver_repos');
SELECT migrate_table('gitserver_repos_statistics');
SELECT migrate_table('gitserver_repos_sync_output');
SELECT migrate_table('global_state');
SELECT migrate_table('insights_query_runner_jobs');
SELECT migrate_table('insights_query_runner_jobs_dependencies');
SELECT migrate_table('insights_settings_migration_jobs');
SELECT migrate_table('lsif_configuration_policies');
SELECT migrate_table('lsif_configuration_policies_repository_pattern_lookup');
SELECT migrate_table('lsif_dependency_indexing_jobs');
SELECT migrate_table('lsif_dependency_repos');
SELECT migrate_table('lsif_dependency_syncing_jobs');
SELECT migrate_table('lsif_dirty_repositories');
SELECT migrate_table('lsif_index_configuration');
SELECT migrate_table('lsif_indexes');
SELECT migrate_table('lsif_last_index_scan');
SELECT migrate_table('lsif_last_retention_scan');
SELECT migrate_table('lsif_nearest_uploads');
SELECT migrate_table('lsif_nearest_uploads_links');
SELECT migrate_table('lsif_packages');
SELECT migrate_table('lsif_references');
SELECT migrate_table('lsif_retention_configuration');
SELECT migrate_table('lsif_uploads');
SELECT migrate_table('lsif_uploads_audit_logs');
SELECT migrate_table('lsif_uploads_reference_counts');
SELECT migrate_table('lsif_uploads_visible_at_tip');
SELECT migrate_table('lsif_uploads_vulnerability_scan');
SELECT migrate_table('names');
SELECT migrate_table('namespace_permissions');
SELECT migrate_table('notebook_stars');
SELECT migrate_table('notebooks');
SELECT migrate_table('org_invitations');
SELECT migrate_table('org_members');
SELECT migrate_table('org_stats');
SELECT migrate_table('orgs');
SELECT migrate_table('orgs_open_beta_stats');
SELECT migrate_table('out_of_band_migrations');
SELECT migrate_table('out_of_band_migrations_errors');
SELECT migrate_table('outbound_webhook_event_types');
SELECT migrate_table('outbound_webhook_jobs');
SELECT migrate_table('outbound_webhook_logs');
SELECT migrate_table('outbound_webhooks');
SELECT migrate_table('own_aggregate_recent_contribution');
SELECT migrate_table('own_aggregate_recent_view');
SELECT migrate_table('own_background_jobs');
SELECT migrate_table('own_signal_configurations');
SELECT migrate_table('own_signal_recent_contribution');
SELECT migrate_table('ownership_path_stats');
SELECT migrate_table('package_repo_filters');
SELECT migrate_table('package_repo_versions');
SELECT migrate_table('permission_sync_jobs');
SELECT migrate_table('permissions');
SELECT migrate_table('phabricator_repos');
SELECT migrate_table('product_licenses');
SELECT migrate_table('product_subscriptions');
SELECT migrate_table('query_runner_state');
SELECT migrate_table('redis_key_value');
SELECT migrate_table('registry_extension_releases');
SELECT migrate_table('registry_extensions');
SELECT migrate_table('repo');
SELECT migrate_table('repo_commits_changelists');
SELECT migrate_table('repo_embedding_job_stats');
SELECT migrate_table('repo_embedding_jobs');
SELECT migrate_table('repo_kvps');
SELECT migrate_table('repo_paths');
SELECT migrate_table('repo_pending_permissions');
SELECT migrate_table('repo_permissions');
SELECT migrate_table('repo_statistics');
SELECT migrate_table('role_permissions');
SELECT migrate_table('roles');
SELECT migrate_table('saved_searches');
SELECT migrate_table('search_context_default');
SELECT migrate_table('search_context_repos');
SELECT migrate_table('search_context_stars');
SELECT migrate_table('search_contexts');
SELECT migrate_table('security_event_logs');
SELECT migrate_table('settings');
SELECT migrate_table('sub_repo_permissions');
SELECT migrate_table('survey_responses');
SELECT migrate_table('syntactic_scip_indexing_jobs');
SELECT migrate_table('syntactic_scip_last_index_scan');
SELECT migrate_table('team_members');
SELECT migrate_table('teams');
SELECT migrate_table('telemetry_events_export_queue');
SELECT migrate_table('temporary_settings');
SELECT migrate_table('user_credentials');
SELECT migrate_table('user_emails');
SELECT migrate_table('user_external_accounts');
SELECT migrate_table('user_onboarding_tour');
SELECT migrate_table('user_pending_permissions');
SELECT migrate_table('user_permissions');
SELECT migrate_table('user_public_repos');
SELECT migrate_table('user_repo_permissions');
SELECT migrate_table('user_roles');
SELECT migrate_table('users');
SELECT migrate_table('versions');
SELECT migrate_table('vulnerabilities');
SELECT migrate_table('vulnerability_affected_packages');
SELECT migrate_table('vulnerability_affected_symbols');
SELECT migrate_table('vulnerability_matches');
SELECT migrate_table('webhook_logs');
SELECT migrate_table('webhooks');
SELECT migrate_table('zoekt_repos');

DROP FUNCTION migrate_table(text);

-- this is required because the id column is hard-overwritten by the inserter
-- and the id number will still be taken by another tenant, those are still globally
-- unique.
alter table out_of_band_migrations alter column id drop default;
ALTER TABLE out_of_band_migrations_errors DROP CONSTRAINT out_of_band_migrations_errors_migration_id_fkey;
ALTER TABLE out_of_band_migrations DROP CONSTRAINT out_of_band_migrations_pkey;
-- also need to make the id, tenant_id unique, instead of just id.
ALTER TABLE out_of_band_migrations ADD PRIMARY KEY (id, tenant_id);
ALTER TABLE out_of_band_migrations_errors ADD CONSTRAINT out_of_band_migrations_errors_migration_id_fkey FOREIGN KEY (migration_id, tenant_id) REFERENCES out_of_band_migrations(id, tenant_id) ON DELETE CASCADE;

-- Need to make unique constraints respect tenant_id.

CREATE OR REPLACE FUNCTION migrate_index(index_name text, table_name text, VARIADIC fields text[])
RETURNS void AS $$
BEGIN
    EXECUTE format('
        ALTER TABLE %I DROP CONSTRAINT IF EXISTS %I;
        ALTER TABLE %I ADD CONSTRAINT %I UNIQUE(%s, tenant_id);',
        table_name, index_name, table_name, index_name, array_to_string(ARRAY(SELECT format('%I', field) FROM unnest(fields) AS field), ', ')
    );
END;
$$ LANGUAGE plpgsql;

-- Existing unique constraints.
SELECT migrate_index('access_requests_email_key', 'access_requests', 'email');
SELECT migrate_index('access_tokens_value_sha256_key', 'access_tokens', 'value_sha256');
SELECT migrate_index('batch_spec_execution_cache_entries_user_id_key_unique', 'batch_spec_execution_cache_entries', 'user_id', 'key');
SELECT migrate_index('batch_spec_resolution_jobs_batch_spec_id_unique', 'batch_spec_resolution_jobs', 'batch_spec_id');
SELECT migrate_index('changeset_events_changeset_id_kind_key_unique', 'changeset_events', 'changeset_id', 'kind', 'key');
SELECT migrate_index('changesets_repo_external_id_unique', 'changesets', 'repo_id', 'external_id');
SELECT migrate_index('code_hosts_url_key', 'code_hosts', 'url');
SELECT migrate_index('codeintel_autoindexing_exceptions_repository_id_key', 'codeintel_autoindexing_exceptions', 'repository_id');
SELECT migrate_index('codeintel_ranking_progress_graph_key_key', 'codeintel_ranking_progress', 'graph_key');
SELECT migrate_index('codeowners_repo_id_key', 'codeowners', 'repo_id');
SELECT migrate_index('executor_heartbeats_hostname_key', 'executor_heartbeats', 'hostname');
SELECT migrate_index('executor_job_tokens_job_id_queue_repo_id_key', 'executor_job_tokens', 'job_id', 'queue', 'repo_id');
SELECT migrate_index('executor_job_tokens_value_sha256_key', 'executor_job_tokens', 'value_sha256');
SELECT migrate_index('external_service_repos_repo_id_external_service_id_unique', 'external_service_repos', 'repo_id', 'external_service_id');
SELECT migrate_index('feature_flag_overrides_unique_org_flag', 'feature_flag_overrides', 'namespace_org_id', 'flag_name');
SELECT migrate_index('feature_flag_overrides_unique_user_flag', 'feature_flag_overrides', 'namespace_user_id', 'flag_name');
SELECT migrate_index('unique_app_install', 'github_app_installs', 'app_id', 'installation_id');
SELECT migrate_index('lsif_index_configuration_repository_id_key', 'lsif_index_configuration', 'repository_id');
SELECT migrate_index('lsif_retention_configuration_repository_id_key', 'lsif_retention_configuration', 'repository_id');
SELECT migrate_index('lsif_uploads_reference_counts_upload_id_key', 'lsif_uploads_reference_counts', 'upload_id');
SELECT migrate_index('org_members_org_id_user_id_key', 'org_members', 'org_id', 'user_id');
SELECT migrate_index('phabricator_repos_repo_name_key', 'phabricator_repos', 'repo_name');
SELECT migrate_index('repo_name_unique', 'repo', 'name');
SELECT migrate_index('repo_pending_permissions_perm_unique', 'repo_pending_permissions', 'repo_id', 'permission');
SELECT migrate_index('repo_permissions_perm_unique', 'repo_permissions', 'repo_id', 'permission');
SELECT migrate_index('search_context_repos_unique', 'search_context_repos', 'repo_id', 'search_context_id', 'revision');
SELECT migrate_index('temporary_settings_user_id_key', 'temporary_settings', 'user_id');
SELECT migrate_index('user_credentials_domain_user_id_external_service_type_exter_key', 'user_credentials', 'domain', 'user_id', 'external_service_type', 'external_service_id');
SELECT migrate_index('user_emails_no_duplicates_per_user', 'user_emails', 'user_id', 'email');
SELECT migrate_index('user_pending_permissions_service_perm_object_unique', 'user_pending_permissions', 'service_type', 'service_id', 'permission', 'object_type', 'bind_id');
SELECT migrate_index('user_permissions_perm_object_unique', 'user_permissions', 'user_id', 'permission', 'object_type');
SELECT migrate_index('user_public_repos_user_id_repo_id_key', 'user_public_repos', 'user_id', 'repo_id');
SELECT migrate_index('webhooks_uuid_key', 'webhooks', 'uuid');

DROP FUNCTION migrate_index(text, text, VARIADIC text[]);

-- Existing unique indexes.


-- TODO: Weird argument ordering.
CREATE OR REPLACE FUNCTION migrate_index(index_name text, where_conds text, table_name text, VARIADIC fields text[])
RETURNS void AS $$
BEGIN
    EXECUTE format('
        DROP INDEX IF EXISTS %I;
        CREATE UNIQUE INDEX IF NOT EXISTS %I ON %I (%s, tenant_id) %s;',
        index_name, index_name, table_name, array_to_string(ARRAY(SELECT format('%I', field) FROM unnest(fields) AS field), ', '), where_conds
    );
END;
$$ LANGUAGE plpgsql;

SELECT migrate_index('assigned_owners_file_path_owner', '', 'assigned_owners', 'file_path_id', 'owner_user_id');
SELECT migrate_index('assigned_teams_file_path_owner', '', 'assigned_teams', 'file_path_id', 'owner_team_id');
SELECT migrate_index('batch_changes_unique_org_id', 'WHERE namespace_org_id IS NOT NULL', 'batch_changes', 'name', 'namespace_org_id');
SELECT migrate_index('batch_changes_unique_user_id', 'WHERE namespace_user_id IS NOT NULL', 'batch_changes', 'name', 'namespace_user_id');
SELECT migrate_index('batch_changes_site_credentials_unique', '', 'batch_changes_site_credentials', 'external_service_type', 'external_service_id');
SELECT migrate_index('batch_spec_workspace_files_batch_spec_id_filename_path', '', 'batch_spec_workspace_files', 'batch_spec_id', 'filename', 'path');
SELECT migrate_index('batch_specs_unique_rand_id', '', 'batch_specs', 'rand_id');
SELECT migrate_index('cached_available_indexers_repository_id', '', 'cached_available_indexers', 'repository_id');
SELECT migrate_index('changeset_specs_unique_rand_id', '', 'changeset_specs', 'rand_id');
SELECT migrate_index('codeintel_autoindex_queue_repository_id_commit', '', 'codeintel_autoindex_queue', 'repository_id', 'rev');
SELECT migrate_index('codeintel_initial_path_ranks_processed_cgraph_key_codeintel_ini', '', 'codeintel_initial_path_ranks_processed', 'graph_key', 'codeintel_initial_path_ranks_id');
SELECT migrate_index('codeintel_langugage_support_requests_user_id_language', '', 'codeintel_langugage_support_requests', 'user_id', 'language_id');
SELECT migrate_index('codeintel_path_ranks_graph_key_repository_id', '', 'codeintel_path_ranks', 'graph_key', 'repository_id');
SELECT migrate_index('codeintel_ranking_exports_graph_key_upload_id', '', 'codeintel_ranking_exports', 'graph_key', 'upload_id');
SELECT migrate_index('codeintel_ranking_path_counts_inputs_graph_key_unique_definitio', 'WHERE NOT processed', 'codeintel_ranking_path_counts_inputs', 'graph_key', 'definition_id');
SELECT migrate_index('codeintel_ranking_references_processed_graph_key_codeintel_rank', '', 'codeintel_ranking_references_processed', 'graph_key', 'codeintel_ranking_reference_id');
SELECT migrate_index('commit_authors_email_name', '', 'commit_authors', 'email', 'name');
SELECT migrate_index('critical_and_site_config_unique', '', 'critical_and_site_config', 'id', 'type');
SELECT migrate_index('event_logs_export_allowlist_event_name_idx', '', 'event_logs_export_allowlist', 'event_name');
SELECT migrate_index('executor_secrets_unique_key_global', 'WHERE namespace_user_id IS NULL AND namespace_org_id IS NULL', 'executor_secrets', 'key', 'scope');
SELECT migrate_index('executor_secrets_unique_key_namespace_org', 'WHERE namespace_org_id IS NOT NULL', 'executor_secrets', 'key', 'namespace_org_id', 'scope');
SELECT migrate_index('executor_secrets_unique_key_namespace_user', 'WHERE namespace_user_id IS NOT NULL', 'executor_secrets', 'key', 'namespace_user_id', 'scope');
SELECT migrate_index('kind_cloud_default', 'WHERE cloud_default = true AND deleted_at IS NULL', 'external_services', 'kind', 'cloud_default');
SELECT migrate_index('github_apps_app_id_slug_base_url_unique', '', 'github_apps', 'app_id', 'slug', 'base_url');
SELECT migrate_index('lsif_dependency_repos_unique_scheme_name', '', 'lsif_dependency_repos', 'scheme', 'name');
SELECT migrate_index('lsif_uploads_repository_id_commit_root_indexer', 'WHERE state = ''completed''::text', 'lsif_uploads', 'repository_id', 'commit', 'root', 'indexer');
SELECT migrate_index('lsif_uploads_vulnerability_scan_upload_id', '', 'lsif_uploads_vulnerability_scan', 'upload_id');
SELECT migrate_index('unique_resource_permission', '', 'namespace_permissions', 'namespace', 'resource_id', 'user_id');
SELECT migrate_index('orgs_name', 'WHERE deleted_at IS NULL', 'orgs', 'name');
SELECT migrate_index('own_aggregate_recent_contribution_file_author', '', 'own_aggregate_recent_contribution', 'changed_file_path_id', 'commit_author_id');
SELECT migrate_index('own_aggregate_recent_view_viewer', '', 'own_aggregate_recent_view', 'viewed_file_path_id', 'viewer_id');
SELECT migrate_index('own_signal_configurations_name_uidx', '', 'own_signal_configurations', 'name');
SELECT migrate_index('package_repo_filters_unique_matcher_per_scheme', '', 'package_repo_filters', 'scheme', 'matcher');
SELECT migrate_index('package_repo_versions_unique_version_per_package', '', 'package_repo_versions', 'package_id', 'version');
SELECT migrate_index('permission_sync_jobs_unique', 'WHERE state = ''queued''::text', 'permission_sync_jobs', 'priority', 'user_id', 'repository_id', 'cancel', 'process_after');
SELECT migrate_index('permissions_unique_namespace_action', '', 'permissions', 'namespace', 'action');
SELECT migrate_index('product_licenses_license_check_token_idx', '', 'product_licenses', 'license_check_token');
SELECT migrate_index('registry_extension_releases_version', 'WHERE release_version IS NOT NULL', 'registry_extension_releases', 'registry_extension_id', 'release_version');
SELECT migrate_index('registry_extensions_publisher_name', 'WHERE deleted_at IS NULL', 'registry_extensions', 'publisher_user_id', 'publisher_org_id', 'name');
SELECT migrate_index('registry_extensions_uuid', '', 'registry_extensions', 'uuid');
SELECT migrate_index('repo_external_unique_idx', '', 'repo', 'external_service_type', 'external_service_id', 'external_id');
SELECT migrate_index('repo_id_perforce_changelist_id_unique', '', 'repo_commits_changelists', 'repo_id', 'perforce_changelist_id');
SELECT migrate_index('repo_paths_index_absolute_path', '', 'repo_paths', 'repo_id', 'absolute_path');
SELECT migrate_index('unique_role_name', '', 'roles', 'name');
SELECT migrate_index('search_contexts_name_namespace_org_id_unique', 'WHERE namespace_org_id IS NOT NULL', 'search_contexts', 'name', 'namespace_org_id');
SELECT migrate_index('search_contexts_name_namespace_user_id_unique', 'WHERE namespace_user_id IS NOT NULL', 'search_contexts', 'name', 'namespace_user_id');
SELECT migrate_index('search_contexts_name_without_namespace_unique', 'WHERE namespace_user_id IS NULL AND namespace_org_id IS NULL', 'search_contexts', 'name');
SELECT migrate_index('sub_repo_permissions_repo_id_user_id_version_uindex', '', 'sub_repo_permissions', 'repo_id', 'user_id', 'version');
SELECT migrate_index('teams_name', '', 'teams', 'name');
SELECT migrate_index('user_emails_user_id_is_primary_idx', 'WHERE is_primary = true', 'user_emails', 'user_id', 'is_primary');
SELECT migrate_index('user_external_accounts_account', 'WHERE deleted_at IS NULL', 'user_external_accounts', 'service_type', 'service_id', 'client_id', 'account_id');
SELECT migrate_index('user_external_accounts_user_id_scim_service_type', 'WHERE service_type = ''scim''::text AND deleted_at IS NULL', 'user_external_accounts', 'user_id', 'service_type');
SELECT migrate_index('user_repo_permissions_perms_unique_idx', '', 'user_repo_permissions', 'user_id', 'user_external_account_id', 'repo_id');
SELECT migrate_index('users_billing_customer_id', 'WHERE deleted_at IS NULL', 'users', 'billing_customer_id');
SELECT migrate_index('users_username', 'WHERE deleted_at IS NULL', 'users', 'username');
SELECT migrate_index('vulnerabilities_source_id', '', 'vulnerabilities', 'source_id');
SELECT migrate_index('vulnerability_affected_packages_vulnerability_id_package_name', '', 'vulnerability_affected_packages', 'vulnerability_id', 'package_name');
SELECT migrate_index('vulnerability_affected_symbols_vulnerability_affected_package_i', '', 'vulnerability_affected_symbols', 'vulnerability_affected_package_id', 'path');
SELECT migrate_index('vulnerability_matches_upload_id_vulnerability_affected_package_', '', 'vulnerability_matches', 'upload_id', 'vulnerability_affected_package_id');


ALTER TABLE syntactic_scip_last_index_scan DROP CONSTRAINT syntactic_scip_last_index_scan_pkey;
ALTER TABLE syntactic_scip_last_index_scan ADD PRIMARY KEY (repository_id, tenant_id);

ALTER TABLE lsif_last_index_scan DROP CONSTRAINT lsif_last_index_scan_pkey;
ALTER TABLE lsif_last_index_scan ADD PRIMARY KEY (repository_id, tenant_id);

ALTER TABLE names DROP CONSTRAINT names_pkey;
ALTER TABLE names ADD PRIMARY KEY (name, tenant_id);

ALTER TABLE user_emails DROP CONSTRAINT user_emails_unique_verified_email;
ALTER TABLE user_emails ADD CONSTRAINT user_emails_unique_verified_email EXCLUDE USING btree(email WITH =, tenant_id WITH =) WHERE (verified_at IS NOT NULL);

DROP FUNCTION migrate_index(text, text, text, VARIADIC text[]);

-- TODO: event_logs_export_allowlist has default values inserted, but not per tenant.
-- todo: for each tenant we have to create system roles:
-- INSERT INTO "public"."roles"("created_at", "system", "name", "tenant_id") VALUES('2023-01-04 17:29:41.195966+01', 'TRUE', 'USER', 2) RETURNING "id", "created_at", "system", "name", "tenant_id";
-- INSERT INTO "public"."roles"("created_at", "system", "name", "tenant_id") VALUES('2023-01-04 17:29:41.195966+01', 'TRUE', 'SITE_ADMINISTRATOR', 2) RETURNING "id", "created_at", "system", "name", "tenant_id";
-- DISABLE_CODE_INSIGHTS: true always needs to be set, too.
