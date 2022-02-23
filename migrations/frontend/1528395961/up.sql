ALTER TABLE IF EXISTS batch_changes
    ALTER COLUMN last_applied_at DROP NOT NULL,
    ALTER COLUMN last_applied_at DROP DEFAULT;
