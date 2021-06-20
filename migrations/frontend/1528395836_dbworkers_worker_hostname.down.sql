BEGIN;

ALTER TABLE changeset_jobs DROP COLUMN IF EXISTS worker_hostname;
ALTER TABLE changesets DROP COLUMN IF EXISTS worker_hostname;
ALTER TABLE cm_action_jobs DROP COLUMN IF EXISTS worker_hostname;
ALTER TABLE cm_trigger_jobs DROP COLUMN IF EXISTS worker_hostname;
ALTER TABLE external_service_sync_jobs DROP COLUMN IF EXISTS worker_hostname;
ALTER TABLE insights_query_runner_jobs DROP COLUMN IF EXISTS worker_hostname;
ALTER TABLE lsif_dependency_indexing_jobs DROP COLUMN IF EXISTS worker_hostname;
ALTER TABLE lsif_indexes DROP COLUMN IF EXISTS worker_hostname;
ALTER TABLE lsif_uploads DROP COLUMN IF EXISTS worker_hostname;

COMMIT;
