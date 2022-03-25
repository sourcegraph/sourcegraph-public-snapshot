DROP TABLE IF EXISTS code_monitors_batch_changes;
ALTER TABLE batch_specs DROP COLUMN IF EXISTS auto_apply;
