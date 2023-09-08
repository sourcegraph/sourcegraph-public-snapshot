-- Undo the changes made in the up migration
DROP INDEX IF EXISTS file_metrics_id_unique;

DROP TABLE IF EXISTS file_metrics;
