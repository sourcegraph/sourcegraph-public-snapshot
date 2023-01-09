ALTER TABLE batch_changes DROP CONSTRAINT IF EXISTS batch_change_name_is_valid;
ALTER TABLE batch_changes ADD CONSTRAINT batch_change_name_is_valid CHECK (name ~ '^[\w.-]+$');
