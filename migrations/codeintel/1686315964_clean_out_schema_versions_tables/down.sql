CREATE TABLE IF NOT EXISTS codeintel_scip_documents_schema_versions (
    upload_id integer PRIMARY KEY,
    min_schema_version integer,
    max_schema_version integer
);
