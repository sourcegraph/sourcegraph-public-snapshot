BEGIN;

INSERT INTO out_of_band_migrations (id, team, component, description, introduced, non_destructive)
VALUES (4, 'code-intelligence', 'codeintel-db.lsif_data_definitions', 'Populate num_locations from gob-encoded payload', '3.26.0', true)
ON CONFLICT DO NOTHING;

INSERT INTO out_of_band_migrations (id, team, component, description, introduced, non_destructive)
VALUES (5, 'code-intelligence', 'codeintel-db.lsif_data_references', 'Populate num_locations from gob-encoded payload', '3.26.0', true)
ON CONFLICT DO NOTHING;

COMMIT;
