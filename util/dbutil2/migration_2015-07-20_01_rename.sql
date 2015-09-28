-- WARNING: Do not run this unless you know what you are doing, this is
-- recorded for historical reasons. Please see the 2015-07-20 section of
-- MIGRATION.md

-- Remove old tables we don't care about
DROP TABLE org_settings;
DROP TABLE person_queue_compute_stats;
DROP TABLE person_queue_refresh_profile;
DROP TABLE person_settings;
DROP TABLE person_stat;
DROP TABLE purged_bids_ended_before_20150130190700_issue_1472;
DROP TABLE repo_backup;
DROP TABLE repo_index;
DROP TABLE repo_index_backup_20141230;
DROP TABLE repo_queue_compute_stats;
DROP TABLE repo_queue_refresh_profile;
DROP TABLE repo_queue_refresh_vcs_data;
DROP TABLE repo_settings;
DROP TABLE repo_settings_backup_20150406;
DROP TABLE repos_that_had_srcbot_enabled_as_of_20141202;
DROP TABLE sqs_successfully_built_repos_20150406;

-- Rename tables that require data migration
ALTER TABLE repo RENAME TO old_repo;
ALTER TABLE repo_build RENAME TO old_repo_build;
ALTER TABLE repo_build_task RENAME TO old_repo_build_task;
ALTER TABLE repo_hit RENAME TO old_repo_hit;
ALTER TABLE repo_key RENAME TO old_repo_key;

-- Rename index on tables
ALTER INDEX repo_pkey RENAME TO old_repo_pkey;
ALTER INDEX repo_name RENAME TO old_repo_name;
ALTER INDEX repo_lower_uri_lower_name RENAME TO old_repo_lower_uri_lower_name;
ALTER INDEX repo_build_pkey RENAME TO old_repo_build_pkey;
ALTER INDEX repo_build_repo RENAME TO old_repo_build_repo;
ALTER INDEX repo_build_priority RENAME TO old_repo_build_priority;
ALTER INDEX repo_build_created_at RENAME TO old_repo_build_created_at;
ALTER INDEX repo_build_updated_at RENAME TO old_repo_build_updated_at;
ALTER INDEX repo_build_successful RENAME TO old_repo_build_successful;
ALTER INDEX repo_build_task_pkey RENAME TO old_repo_build_task_pkey;
ALTER INDEX repo_build_task_build RENAME TO old_repo_build_task_build;
ALTER INDEX repo_hit_repo RENAME TO old_repo_hit_repo;
ALTER INDEX repo_hit_repo_at RENAME TO old_repo_hit_repo_at;
ALTER INDEX repo_key_pkey RENAME TO old_repo_key_pkey;
