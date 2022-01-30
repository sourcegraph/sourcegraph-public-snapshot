BEGIN;

ALTER TABLE IF EXISTS batch_spec_resolution_jobs
  ADD CONSTRAINT batch_spec_resolution_jobs_batch_spec_id_unique UNIQUE (batch_spec_id);

COMMIT;
