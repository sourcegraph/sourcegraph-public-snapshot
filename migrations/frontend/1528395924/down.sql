BEGIN;

-- Drop new table
DROP TABLE IF EXISTS lsif_configuration_policies_repository_pattern_lookup;

-- Drop new columns
ALTER TABLE lsif_configuration_policies DROP COLUMN last_resolved_at;
ALTER TABLE lsif_configuration_policies DROP COLUMN repository_patterns;

COMMIT;
