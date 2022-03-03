UPDATE batch_changes SET last_applied_at = TO_TIMESTAMP(0) WHERE last_applied_at IS NULL;

ALTER TABLE IF EXISTS batch_changes
    ALTER COLUMN last_applied_at SET DEFAULT NOW(),
    ALTER COLUMN last_applied_at SET NOT NULL;
