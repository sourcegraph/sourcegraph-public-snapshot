BEGIN;

INSERT INTO out_of_band_migrations (id, team, component, description, introduced, non_destructive)
VALUES (1, 'code-intelligence', 'codeintel-db.lsif_data_documents', 'Populate num_diagnostics from gob-encoded payload', '3.25.0', true)
ON CONFLICT DO NOTHING;

COMMIT;
