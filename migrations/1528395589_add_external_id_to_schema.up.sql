BEGIN;
ALTER TABLE changesets ADD COLUMN external_id text
  NOT NULL CHECK (external_id != '');
ALTER TABLE changesets ADD CONSTRAINT changesets_repo_external_id_unique
  UNIQUE (repo_id, external_id);
COMMIT;
