ALTER TABLE batch_specs
    ADD COLUMN IF NOT EXISTS batch_change_id bigint,
    ADD FOREIGN KEY (batch_change_id) REFERENCES batch_changes(id) ON DELETE SET NULL DEFERRABLE;

UPDATE batch_specs SET batch_change_id = (
    -- In the event, the database is in a not-so-optimal state (usually in dev environment)
    -- we want the subquery to never return more than one row.
    -- This will never happen in production.
    SELECT
        bc.id
    FROM batch_changes bc
    WHERE bc.batch_spec_id = batch_specs.id
    LIMIT 1
);
