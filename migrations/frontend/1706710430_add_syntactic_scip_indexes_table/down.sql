-- Undo the changes made in the up migration

DROP TABLE IF EXISTS syntactic_scip_indexes;
DROP VIEW IF EXISTS syntactic_scip_indexes_with_repository_name;
DROP TABLE CASCADE IF EXISTS syntactic_scip_indexes;
