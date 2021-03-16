BEGIN;

ALTER TABLE changesets ADD COLUMN IF NOT EXISTS external_title text;
COMMENT ON COLUMN changesets.external_title IS 'Normalized property generated on save using Changeset.Title()';

UPDATE changesets SET external_title = COALESCE(changesets.metadata->>'Title', changesets.metadata->>'title', NULL);

CREATE INDEX IF NOT EXISTS changesets_external_title_idx ON changesets USING BTREE(external_title);

COMMIT;
