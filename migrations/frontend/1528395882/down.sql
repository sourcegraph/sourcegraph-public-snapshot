BEGIN;

ALTER TABLE lsif_configuration_policies DROP COLUMN protected;

COMMIT;
