DROP INDEX IF EXISTS codeintel_lockfile_references_repository_id_commit_bytea;

CREATE INDEX IF NOT EXISTS codeintel_lockfile_references_repository_id_commit_bytea ON codeintel_lockfile_references USING btree (repository_id, commit_bytea)
WHERE
    repository_id IS NOT NULL
    AND commit_bytea IS NOT NULL;

ALTER TABLE
    codeintel_lockfile_references
ADD
    COLUMN IF NOT EXISTS last_check_at timestamptz;

COMMENT ON COLUMN codeintel_lockfile_references.last_check_at IS 'Timestamp when background job last checked this row for repository resolution';

CREATE INDEX IF NOT EXISTS codeintel_lockfile_references_last_check_at ON codeintel_lockfile_references USING btree (last_check_at);
