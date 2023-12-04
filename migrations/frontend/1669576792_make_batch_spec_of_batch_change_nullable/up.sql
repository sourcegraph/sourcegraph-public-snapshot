ALTER TABLE batch_changes ALTER COLUMN batch_spec_id DROP NOT NULL;

CREATE TEMPORARY TABLE minimal_batch_specs (id bigint);

INSERT INTO minimal_batch_specs
SELECT batch_spec_id FROM batch_changes WHERE batch_spec_id IS NOT NULL AND last_applied_at IS NULL;

UPDATE batch_changes SET batch_spec_id = NULL WHERE batch_spec_id IN (SELECT id FROM minimal_batch_specs);

-- Delete existing empty batch specs.
DELETE FROM batch_specs WHERE id IN (SELECT id FROM minimal_batch_specs);-- Delete existing empty batch specs.

DROP TABLE minimal_batch_specs;
