-- +++
-- parent: 1528395968
-- +++

BEGIN;

ALTER TABLE batch_spec_workspaces DROP COLUMN IF EXISTS steps;
ALTER TABLE batch_spec_workspaces DROP COLUMN IF EXISTS skipped_steps;

COMMIT;
