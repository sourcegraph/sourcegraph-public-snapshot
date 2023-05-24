ALTER TABLE executor_heartbeats ALTER COLUMN queue_name DROP NOT NULL;
ALTER TABLE executor_heartbeats ADD COLUMN IF NOT EXISTS queue_names TEXT[];
ALTER TABLE executor_heartbeats ADD CONSTRAINT queue_not_null CHECK (
    (queue_name IS NOT NULL AND queue_names IS NULL)
    OR
    (queue_names IS NOT NULL AND queue_name IS NULL)
)
