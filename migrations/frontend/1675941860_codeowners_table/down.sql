-- Undo the changes made in the up migration
DROP VIEW IF EXISTS own_blame_jobs_with_repository_name;
DROP TABLE IF EXISTS own_blame_jobs;
DROP TABLE IF EXISTS own_signals;
DROP TABLE IF EXISTS own_artifacts;
