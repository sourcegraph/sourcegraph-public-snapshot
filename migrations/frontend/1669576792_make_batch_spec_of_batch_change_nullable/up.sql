ALTER TABLE batch_changes ALTER COLUMN batch_spec_id DROP NOT NULL;

-- Delete existing empty batch specs.
DELETE FROM batch_specs WHERE id IN (SELECT batch_spec_id FROM batch_changes WHERE batch_spec_id IS NOT NULL AND last_applied_at IS NULL);-- Delete existing empty batch specs.
