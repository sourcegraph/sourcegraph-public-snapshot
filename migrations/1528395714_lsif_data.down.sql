BEGIN;

DROP TABLE lsif_data_metadata CASCADE;
DROP TABLE lsif_data_documents CASCADE;
DROP TABLE lsif_data_result_chunks CASCADE;
DROP TABLE lsif_data_definitions CASCADE;
DROP TABLE lsif_data_references CASCADE;

DROP SERVER IF EXISTS shard1 CASCADE;
DROP SERVER IF EXISTS shard2 CASCADE;

COMMIT;
