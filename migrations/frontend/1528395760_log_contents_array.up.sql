BEGIN;

-- Drop dependent views
DROP VIEW lsif_indexes_with_repository_name;
DROP VIEW external_service_sync_jobs_with_next_sync_at;

-- Create new columns
ALTER TABLE lsif_indexes ADD COLUMN execution_logs json[];
ALTER TABLE changesets ADD COLUMN execution_logs json[];
ALTER TABLE external_service_sync_jobs ADD execution_logs json[];

-- Back-fill log data into new column
UPDATE lsif_indexes SET execution_logs = (execution_logs || json_build_object('command', '{}'::text[], 'out', log_contents)) WHERE log_contents IS NOT NULL;

--
-- Recreate views with new columns

CREATE VIEW lsif_indexes_with_repository_name AS
    SELECT u.*, r.name as repository_name FROM lsif_indexes u
    JOIN repo r ON r.id = u.repository_id
    WHERE r.deleted_at IS NULL;

CREATE VIEW external_service_sync_jobs_with_next_sync_at AS SELECT
    j.id,
    j.state,
    j.failure_message,
    j.started_at,
    j.finished_at,
    j.process_after,
    j.num_resets,
    j.num_failures,
    j.execution_logs,
    j.external_service_id,
    e.next_sync_at
FROM external_services e JOIN external_service_sync_jobs j ON e.id = j.external_service_id;

COMMIT;
