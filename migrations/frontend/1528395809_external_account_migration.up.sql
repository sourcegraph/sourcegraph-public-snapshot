BEGIN;

INSERT INTO out_of_band_migrations (id, team, component, description, introduced, non_destructive)
VALUES (6, 'core-application', 'frontend-db.external-accounts', 'Encrypt auth data', '3.26.0', true)
ON CONFLICT DO NOTHING;

COMMIT;
