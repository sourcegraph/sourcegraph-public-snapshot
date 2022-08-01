ALTER TABLE batch_specs
    ADD COLUMN IF NOT EXISTS batch_change_id bigint,
    ADD FOREIGN KEY (batch_change_id) REFERENCES batch_changes(id) ON DELETE SET NULL DEFERRABLE;

UPDATE batch_specs SET batch_change_id = (
    -- In the event that a batch change is manually mapped to different batch specs,
    -- we want to use the oldest. This should never happen in production,
    -- it's only added to minimize errors when running the migration on dev.
    SELECT
        bc.id
    FROM batch_changes bc
    WHERE bc.batch_spec_id = batch_specs.id
    LIMIT 1
);
