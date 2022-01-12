-- +++
-- parent: 1528395917
-- +++

BEGIN;

ALTER TABLE batch_spec_workspaces
  ADD COLUMN IF NOT EXISTS ignored BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS unsupported BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS skipped BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE batch_specs
  ADD COLUMN IF NOT EXISTS allow_unsupported BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS allow_ignored BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE batch_spec_resolution_jobs
  DROP COLUMN IF EXISTS allow_unsupported,
  DROP COLUMN IF EXISTS allow_ignored;

COMMIT;
