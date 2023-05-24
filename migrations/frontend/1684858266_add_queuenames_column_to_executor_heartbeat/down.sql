-- multi-queue executors are not implemented in the version that this down migration reverts to, so
-- delete all heartbeats that contain multiple queue names due to queue_name being NULL for those rows
DELETE FROM executor_heartbeats WHERE queue_name IS NULL AND queue_names IS NOT NULL;

ALTER TABLE executor_heartbeats DROP CONSTRAINT IF EXISTS queue_not_null;
ALTER TABLE executor_heartbeats DROP COLUMN IF EXISTS queue_names;
ALTER TABLE executor_heartbeats ALTER COLUMN queue_name SET NOT NULL;
