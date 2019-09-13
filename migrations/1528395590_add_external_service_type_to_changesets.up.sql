BEGIN;

ALTER TABLE changesets ADD COLUMN external_service_type text;

UPDATE changesets
SET
  external_service_type = repo.external_service_type
FROM changesets cs
JOIN
  repo
ON
  repo.id = cs.repo_id;

ALTER TABLE changesets
  ALTER COLUMN external_service_type SET NOT NULL;

ALTER TABLE changesets
  ADD CONSTRAINT changesets_external_service_type_not_blank
  CHECK (external_service_type != '');

COMMIT;
