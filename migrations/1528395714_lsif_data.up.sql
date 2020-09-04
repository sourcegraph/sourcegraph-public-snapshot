BEGIN;

--
-- In devleopment:
--
-- docker run --rm -d -e POSTGRES_USER=sg -e POSTGRES_PASSWORD=sg -p 5433:5432 postgres:12.4 # shard 1
-- docker run --rm -d -e POSTGRES_USER=sg -e POSTGRES_PASSWORD=sg -p 5434:5432 postgres:12.4 # shard 2
--

--
-- Before applying this migration, run the following in shard1 and shard2
--
-- CREATE TABLE lsif_data_metadata (dump_id int, num_result_chunks int, PRIMARY KEY (dump_id));
-- CREATE TABLE lsif_data_documents (dump_id int, path text, data bytea, PRIMARY KEY (dump_id, path));
-- CREATE TABLE lsif_data_result_chunks (dump_id int, idx int, data bytea, PRIMARY KEY (dump_id, idx));
-- CREATE TABLE lsif_data_definitions (dump_id int, scheme text, identifier text, data bytea, PRIMARY KEY (dump_id, scheme, identifier));
-- CREATE TABLE lsif_data_references (dump_id int, scheme text, identifier text, data bytea, PRIMARY KEY (dump_id, scheme, identifier));
--

CREATE EXTENSION IF NOT EXISTS postgres_fdw;

-- CREATE SERVER IF NOT EXISTS shard1 FOREIGN DATA WRAPPER postgres_fdw OPTIONS (dbname 'sg', host 'localhost', port '5433');
-- CREATE USER MAPPING FOR sg SERVER shard1 OPTIONS (user 'sg', password 'sg');

-- CREATE SERVER IF NOT EXISTS shard2 FOREIGN DATA WRAPPER postgres_fdw OPTIONS (dbname 'sg', host 'localhost', port '5434');
-- CREATE USER MAPPING FOR sg SERVER shard2 OPTIONS (user 'sg', password 'sg');

-- -- Sharded metadata table
-- CREATE TABLE lsif_data_metadata (dump_id int, num_result_chunks int) PARTITION BY HASH(dump_id);
-- CREATE FOREIGN TABLE lsif_data_metadata_0 PARTITION OF lsif_data_metadata FOR VALUES WITH (modulus 2, remainder 0) SERVER shard1 OPTIONS (table_name 'lsif_data_metadata');
-- CREATE FOREIGN TABLE lsif_data_metadata_1 PARTITION OF lsif_data_metadata FOR VALUES WITH (modulus 2, remainder 1) SERVER shard2 OPTIONS (table_name 'lsif_data_metadata');

-- -- Sharded documents table
-- CREATE TABLE lsif_data_documents (dump_id int, path text, data bytea) PARTITION BY HASH(dump_id);
-- CREATE FOREIGN TABLE lsif_data_documents_0 PARTITION OF lsif_data_documents FOR VALUES WITH (modulus 2, remainder 0) SERVER shard1 OPTIONS (table_name 'lsif_data_documents');
-- CREATE FOREIGN TABLE lsif_data_documents_1 PARTITION OF lsif_data_documents FOR VALUES WITH (modulus 2, remainder 1) SERVER shard2 OPTIONS (table_name 'lsif_data_documents');

-- -- Sharded result chunks table
-- CREATE TABLE lsif_data_result_chunks (dump_id int, idx int, data bytea) PARTITION BY HASH(dump_id);
-- CREATE FOREIGN TABLE lsif_data_result_chunks_0 PARTITION OF lsif_data_result_chunks FOR VALUES WITH (modulus 2, remainder 0) SERVER shard1 OPTIONS (table_name 'lsif_data_result_chunks');
-- CREATE FOREIGN TABLE lsif_data_result_chunks_1 PARTITION OF lsif_data_result_chunks FOR VALUES WITH (modulus 2, remainder 1) SERVER shard2 OPTIONS (table_name 'lsif_data_result_chunks');

-- -- Sharded definitions table
-- CREATE TABLE lsif_data_definitions (dump_id int, scheme text, identifier text, data bytea) PARTITION BY HASH(dump_id);
-- CREATE FOREIGN TABLE lsif_data_definitions_0 PARTITION OF lsif_data_definitions FOR VALUES WITH (modulus 2, remainder 0) SERVER shard1 OPTIONS (table_name 'lsif_data_definitions');
-- CREATE FOREIGN TABLE lsif_data_definitions_1 PARTITION OF lsif_data_definitions FOR VALUES WITH (modulus 2, remainder 1) SERVER shard2 OPTIONS (table_name 'lsif_data_definitions');

-- -- Sharded references table
-- CREATE TABLE lsif_data_references (dump_id int, scheme text, identifier text, data bytea) PARTITION BY HASH(dump_id);
-- CREATE FOREIGN TABLE lsif_data_references_0 PARTITION OF lsif_data_references FOR VALUES WITH (modulus 2, remainder 0) SERVER shard1 OPTIONS (table_name 'lsif_data_references');
-- CREATE FOREIGN TABLE lsif_data_references_1 PARTITION OF lsif_data_references FOR VALUES WITH (modulus 2, remainder 1) SERVER shard2 OPTIONS (table_name 'lsif_data_references');


CREATE TABLE lsif_data_metadata (dump_id int, num_result_chunks int, PRIMARY KEY (dump_id));
CREATE TABLE lsif_data_documents (dump_id int, path text, data bytea, PRIMARY KEY (dump_id, path));
CREATE TABLE lsif_data_result_chunks (dump_id int, idx int, data bytea, PRIMARY KEY (dump_id, idx));
CREATE TABLE lsif_data_definitions (dump_id int, scheme text, identifier text, data bytea, PRIMARY KEY (dump_id, scheme, identifier));
CREATE TABLE lsif_data_references (dump_id int, scheme text, identifier text, data bytea, PRIMARY KEY (dump_id, scheme, identifier));

COMMIT;
