-- +++
-- parent: 1528395872
-- +++

BEGIN;

INSERT INTO out_of_band_migrations (id, team, component, description, introduced_version_major, introduced_version_minor, non_destructive)
VALUES (11, 'code-intelligence', 'lsif_uploads.num_references', 'Backfill LSIF upload reference counts', 3, 22, true)
ON CONFLICT DO NOTHING;

COMMIT;
