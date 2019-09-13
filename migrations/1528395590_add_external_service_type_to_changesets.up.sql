BEGIN;

ALTER TABLE changesets ADD COLUMN external_service_type text;

UPDATE changesets SET external_service_type = 'github';

ALTER TABLE changesets
  ADD CONSTRAINT changesets_external_service_type_not_null
  CHECK (external_service_type IS NOT NULL);

ALTER TABLE changesets
  ADD CONSTRAINT changesets_external_service_type_not_blank
  CHECK (external_service_type != '');

COMMIT;
