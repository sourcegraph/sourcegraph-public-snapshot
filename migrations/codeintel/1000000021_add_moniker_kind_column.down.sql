BEGIN;

ALTER TABLE lsif_data_references DROP COLUMN kind;
ALTER TABLE lsif_data_definitions DROP COLUMN kind;

COMMIT;
