BEGIN;

ALTER TABLE batch_spec_workspaces
  DROP COLUMN IF EXISTS ignored,
  DROP COLUMN IF EXISTS unsupported;

COMMIT;
