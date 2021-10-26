BEGIN;

ALTER TABLE lsif_configuration_policies DROP COLUMN repository_patterns;

COMMIT;
