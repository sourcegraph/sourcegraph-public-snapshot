BEGIN;

INSERT INTO out_of_band_migrations (id, team, component, description, introduced, deprecated, non_destructive)
VALUES (2, 'campaigns', 'frontend-db.authenticators', 'Prepare for SSH pushes to code hosts', '3.26.0', '3.27.0', true)
ON CONFLICT DO NOTHING;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
