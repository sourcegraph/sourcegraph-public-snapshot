BEGIN;

CREATE TABLE lsif_data_metadata (dump_id int, num_result_chunks int, PRIMARY KEY (dump_id));
CREATE TABLE lsif_data_documents (dump_id int, path text, data bytea, PRIMARY KEY (dump_id, path));
CREATE TABLE lsif_data_result_chunks (dump_id int, idx int, data bytea, PRIMARY KEY (dump_id, idx));
CREATE TABLE lsif_data_definitions (dump_id int, scheme text, identifier text, data bytea, PRIMARY KEY (dump_id, scheme, identifier));
CREATE TABLE lsif_data_references (dump_id int, scheme text, identifier text, data bytea, PRIMARY KEY (dump_id, scheme, identifier));

COMMIT;
