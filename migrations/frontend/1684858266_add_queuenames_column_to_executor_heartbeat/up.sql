ALTER TABLE executor_heartbeats ALTER COLUMN queue_name DROP NOT NULL;
ALTER TABLE executor_heartbeats ADD COLUMN IF NOT EXISTS queue_names TEXT[];

DO $$
BEGIN
    IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conname = 'one_of_queue_name_queue_names') THEN
        ALTER TABLE executor_heartbeats
            ADD CONSTRAINT one_of_queue_name_queue_names
                CHECK (
                        (queue_name IS NOT NULL AND queue_names IS NULL)
                        OR
                        (queue_names IS NOT NULL AND queue_name IS NULL)
                    );
    END IF;
END $$;

COMMENT ON COLUMN executor_heartbeats.queue_names IS 'The list of queue names that the executor polls for work.';
