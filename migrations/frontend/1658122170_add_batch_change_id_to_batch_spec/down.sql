-- Undo the changes made in the up migration
ALTER TABLE batch_specs
    DROP COLUMN IF EXISTS batch_change_id;
