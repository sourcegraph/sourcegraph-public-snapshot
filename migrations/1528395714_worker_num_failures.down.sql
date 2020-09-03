BEGIN;

ALTER TABLE lsif_uploads DROP COLUMN num_failures;
ALTER TABLE lsif_indexes DROP COLUMN num_failures;
ALTER TABLE changesets DROP COLUMN num_failures;
ALTER TABLE external_service_sync_jobs DROP COLUMN num_failures;

COMMIT;
