BEGIN;

ALTER TABLE IF EXISTS batch_spec_resolution_jobs
  DROP CONSTRAINT IF EXISTS batch_spec_resolution_jobs_batch_spec_id_unique;

COMMIT;
