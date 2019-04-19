BEGIN;
UPDATE external_services SET config = '{}' WHERE BTRIM(config) = '';
ALTER TABLE external_services ADD CONSTRAINT check_non_empty_config CHECK (BTRIM(config) <> '');
COMMIT;
