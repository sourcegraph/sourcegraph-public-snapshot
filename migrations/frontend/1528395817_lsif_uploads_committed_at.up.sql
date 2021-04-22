BEGIN;

ALTER TABLE lsif_uploads ADD COLUMN committed_at timestamp with time zone;
CREATE INDEX lsif_uploads_committed_at ON lsif_uploads (committed_at) WHERE state = 'completed';

INSERT INTO out_of_band_migrations (id, team, component, description, introduced, non_destructive)
VALUES (8, 'code-intelligence', 'frontend-db.lsif_uploads', 'Backfill committed_at', '3.28.0', true)
ON CONFLICT DO NOTHING;

COMMIT;
