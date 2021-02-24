BEGIN;

INSERT INTO out_of_band_migrations (id, team, component, description, introduced, deprecated, non_destructive)
VALUES (2, 'campaigns', 'frontend-db.authenticators', 'Prepare for SSH pushes to code hosts', '3.26.0', '3.27.0', true)
ON CONFLICT DO NOTHING;

COMMIT;
