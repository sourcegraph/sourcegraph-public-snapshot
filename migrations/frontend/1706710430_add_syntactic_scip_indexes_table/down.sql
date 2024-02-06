-- Undo the changes made in the up migration

DROP VIEW IF EXISTS syntactic_scip_indexes_with_repository_name;
DROP TABLE IF EXISTS syntactic_scip_indexes;
