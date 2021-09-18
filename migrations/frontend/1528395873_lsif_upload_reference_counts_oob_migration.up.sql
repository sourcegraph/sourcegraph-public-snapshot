BEGIN;

INSERT INTO out_of_band_migrations (id, team, component, description, introduced_version_major, introduced_version_minor, non_destructive)
VALUES (11, 'code-intelligence', 'lsif_uploads.num_references', 'Backfill LSIF upload reference counts', 3, 22, true)
ON CONFLICT DO NOTHING;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
