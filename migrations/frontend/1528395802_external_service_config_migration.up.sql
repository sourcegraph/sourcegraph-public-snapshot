BEGIN;

INSERT INTO out_of_band_migrations (id, team, component, description, introduced, deprecated, non_destructive)
VALUES (3, 'core-application', 'frontend-db.external-services', 'Encrypt configuration', '3.26.0', '3.27.0', true)
ON CONFLICT DO NOTHING;

COMMIT;
