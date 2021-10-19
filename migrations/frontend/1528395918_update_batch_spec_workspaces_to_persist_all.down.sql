BEGIN;

ALTER TABLE batch_spec_workspaces
  DROP COLUMN IF EXISTS ignored,
  DROP COLUMN IF EXISTS unsupported,
  DROP COLUMN IF EXISTS skipped;

ALTER TABLE batch_specs
  DROP COLUMN IF EXISTS allow_unsupported,
  DROP COLUMN IF EXISTS allow_ignored;

ALTER TABLE batch_spec_resolution_jobs
  ADD COLUMN IF NOT EXISTS allow_unsupported BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS allow_ignored BOOLEAN NOT NULL DEFAULT FALSE;

COMMIT;
