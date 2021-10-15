BEGIN;

ALTER TABLE batch_spec_workspaces
  DROP COLUMN IF EXISTS skipped;

COMMIT;
