BEGIN;

-- Drop new column
ALTER TABLE lsif_configuration_policies DROP COLUMN last_resolved_at;

-- Drop new table
DROP TABLE IF EXISTS lsif_configuration_policies_repository_pattern_lookup;

COMMIT;
