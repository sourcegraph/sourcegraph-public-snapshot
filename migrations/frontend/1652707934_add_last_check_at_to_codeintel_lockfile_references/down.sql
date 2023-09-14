ALTER TABLE
    codeintel_lockfile_references
DROP
    COLUMN IF EXISTS last_check_at;

DROP INDEX codeintel_lockfile_references_repository_id_commit_bytea;
CREATE UNIQUE INDEX codeintel_lockfile_references_repository_id_commit_bytea
ON codeintel_lockfile_references
USING btree (repository_id, commit_bytea)
WHERE repository_id IS NOT NULL AND commit_bytea IS NOT NULL;
