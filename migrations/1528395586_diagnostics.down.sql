BEGIN;

ALTER TABLE events DROP COLUMN thread_diagnostic_edge_id;
DROP TABLE IF EXISTS thread_diagnostic_edges;

COMMIT;
