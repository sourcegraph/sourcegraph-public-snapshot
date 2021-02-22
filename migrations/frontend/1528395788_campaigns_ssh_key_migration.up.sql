BEGIN;

INSERT INTO out_of_band_migrations (id, team, component, description, introduced, non_destructive)
VALUES (2, 'campaigns', 'campaigns-db.authenticators', 'Prepare for SSH pushes to code hosts', '3.26.0', true)
ON CONFLICT DO NOTHING;

COMMIT;
