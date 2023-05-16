DELETE INDEX IF EXISTS perforce_changelist_mapping_jobs_id_repo_id_unique;

DELETE INDEX IF EXISTS perforce_changelist_mapping_jobs_state;

DELETE INDEX IF EXISTS perforce_changelist_mapping_jobs_process_after;

DROP TABLE IF EXISTS perforce_changelist_mapping_jobs;
