BEGIN;

ALTER TABLE changeset_jobs ADD COLUMN worker_hostname text NOT NULL DEFAULT '';
ALTER TABLE changesets ADD COLUMN IF NOT EXISTS worker_hostname text NOT NULL DEFAULT '';
ALTER TABLE cm_action_jobs ADD COLUMN IF NOT EXISTS worker_hostname text NOT NULL DEFAULT '';
ALTER TABLE cm_trigger_jobs ADD COLUMN IF NOT EXISTS worker_hostname text NOT NULL DEFAULT '';
ALTER TABLE external_service_sync_jobs ADD COLUMN IF NOT EXISTS worker_hostname text NOT NULL DEFAULT '';
ALTER TABLE insights_query_runner_jobs ADD COLUMN IF NOT EXISTS worker_hostname text NOT NULL DEFAULT '';
ALTER TABLE lsif_dependency_indexing_jobs ADD COLUMN IF NOT EXISTS worker_hostname text NOT NULL DEFAULT '';
ALTER TABLE lsif_indexes ADD COLUMN IF NOT EXISTS worker_hostname text NOT NULL DEFAULT '';
ALTER TABLE lsif_uploads ADD COLUMN IF NOT EXISTS worker_hostname text NOT NULL DEFAULT '';

COMMIT;
