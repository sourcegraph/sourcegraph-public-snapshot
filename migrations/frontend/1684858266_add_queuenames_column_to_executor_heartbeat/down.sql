-- multi-queue executors are not implemented in the version that this down migration reverts to, so
-- delete all heartbeats that contain multiple queue names due to queue_name being NULL for those rows
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'executor_heartbeats' AND column_name = 'queue_names') THEN
        DELETE FROM executor_heartbeats WHERE queue_name IS NULL AND queue_names IS NOT NULL;
    END IF;
END $$;

ALTER TABLE executor_heartbeats DROP CONSTRAINT IF EXISTS one_of_queue_name_queue_names;
ALTER TABLE executor_heartbeats DROP COLUMN IF EXISTS queue_names;
ALTER TABLE executor_heartbeats ALTER COLUMN queue_name SET NOT NULL;
