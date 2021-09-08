BEGIN;

DROP TABLE IF EXISTS batch_spec_workspace_executions;

ALTER TABLE IF EXISTS batch_specs
  DROP COLUMN IF EXISTS state,
  DROP COLUMN IF EXISTS failure_message,
  DROP COLUMN IF EXISTS started_at,
  DROP COLUMN IF EXISTS finished_at,
  DROP COLUMN IF EXISTS process_after,
  DROP COLUMN IF EXISTS num_resets,
  DROP COLUMN IF EXISTS num_failures;

COMMIT;
