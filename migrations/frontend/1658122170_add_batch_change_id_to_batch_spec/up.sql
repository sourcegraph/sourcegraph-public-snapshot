ALTER TABLE batch_specs
    ADD COLUMN IF NOT EXISTS batch_change_id bigint,
    ADD FOREIGN KEY (batch_change_id) REFERENCES batch_changes(id) ON DELETE SET NULL DEFERRABLE;

UPDATE batch_specs SET batch_change_id = (SELECT bc.id FROM batch_changes bc WHERE bc.batch_spec_id = batch_specs.id);
