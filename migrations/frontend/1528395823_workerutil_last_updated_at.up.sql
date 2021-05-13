BEGIN;

ALTER TABLE changeset_jobs ADD COLUMN last_updated_at timestamp with time zone;
ALTER TABLE changesets ADD COLUMN last_updated_at timestamp with time zone;
ALTER TABLE cm_action_jobs ADD COLUMN last_updated_at timestamp with time zone;
ALTER TABLE cm_trigger_jobs ADD COLUMN last_updated_at timestamp with time zone;
ALTER TABLE external_service_sync_jobs ADD COLUMN last_updated_at timestamp with time zone;
ALTER TABLE insights_query_runner_jobs ADD COLUMN last_updated_at timestamp with time zone;
ALTER TABLE lsif_indexes ADD COLUMN last_updated_at timestamp with time zone;
ALTER TABLE lsif_uploads ADD COLUMN last_updated_at timestamp with time zone;

COMMIT;
