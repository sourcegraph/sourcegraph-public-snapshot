BEGIN;

-- Drop dependent views
DROP VIEW lsif_dumps_with_repository_name;
DROP VIEW lsif_indexes_with_repository_name;
DROP VIEW lsif_uploads_with_repository_name;
DROP VIEW lsif_dumps;

-- Drop columns
ALTER TABLE lsif_uploads DROP COLUMN num_failures;
ALTER TABLE lsif_indexes DROP COLUMN num_failures;
ALTER TABLE changesets DROP COLUMN num_failures;
ALTER TABLE external_service_sync_jobs DROP COLUMN num_failures;

-- Recreate views
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

CREATE VIEW lsif_dumps_with_repository_name AS
    SELECT u.*, r.name as repository_name FROM lsif_dumps u
    JOIN repo r ON r.id = u.repository_id
    WHERE r.deleted_at IS NULL;

CREATE VIEW lsif_uploads_with_repository_name AS
    SELECT u.*, r.name as repository_name FROM lsif_uploads u
    JOIN repo r ON r.id = u.repository_id
    WHERE r.deleted_at IS NULL;

CREATE VIEW lsif_indexes_with_repository_name AS
    SELECT u.*, r.name as repository_name FROM lsif_indexes u
    JOIN repo r ON r.id = u.repository_id
    WHERE r.deleted_at IS NULL;

COMMIT;
