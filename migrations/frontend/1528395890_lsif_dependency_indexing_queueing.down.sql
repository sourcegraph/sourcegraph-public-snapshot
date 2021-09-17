
BEGIN;

DROP TABLE IF EXISTS lsif_dependency_indexing_jobs;

ALTER TABLE lsif_dependency_syncing_jobs
RENAME TO lsif_dependency_indexing_jobs;

COMMIT;
