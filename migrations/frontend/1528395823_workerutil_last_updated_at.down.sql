BEGIN;

ALTER TABLE changeset_jobs DROP COLUMN last_updated_at;
ALTER TABLE changesets DROP COLUMN last_updated_at;
ALTER TABLE cm_action_jobs DROP COLUMN last_updated_at;
ALTER TABLE cm_trigger_jobs DROP COLUMN last_updated_at;
ALTER TABLE external_service_sync_jobs DROP COLUMN last_updated_at;
ALTER TABLE insights_query_runner_jobs DROP COLUMN last_updated_at;
ALTER TABLE lsif_indexes DROP COLUMN last_updated_at;
ALTER TABLE lsif_uploads DROP COLUMN last_updated_at;

COMMIT;
