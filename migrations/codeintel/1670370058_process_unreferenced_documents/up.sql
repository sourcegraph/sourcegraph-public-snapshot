CREATE TABLE IF NOT EXISTS codeintel_scip_documents_dereference_logs (
    id bigserial,
    document_id bigint NOT NULL,
    last_removal_time timestamp with time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY(id)
);

COMMENT ON TABLE codeintel_scip_documents_dereference_logs IS 'A list of document rows that were recently dereferenced by the deletion of an index.';
COMMENT ON COLUMN codeintel_scip_documents_dereference_logs.document_id IS 'The identifier of the document that was dereferenced.';
COMMENT ON COLUMN codeintel_scip_documents_dereference_logs.last_removal_time IS 'The time that the log entry was inserted.';

CREATE INDEX IF NOT EXISTS codeintel_scip_documents_dereference_logs_last_removal_time_document_id ON codeintel_scip_documents_dereference_logs(last_removal_time, document_id);

CREATE OR REPLACE FUNCTION update_codeintel_scip_documents_dereference_logs_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
    INSERT INTO codeintel_scip_documents_dereference_logs (document_id)
    SELECT document_id FROM oldtab;
    RETURN NULL;
END $$;

DROP TRIGGER IF EXISTS codeintel_scip_documents_dereference_logs_insert ON codeintel_scip_document_lookup;
CREATE TRIGGER codeintel_scip_documents_dereference_logs_insert AFTER DELETE ON codeintel_scip_document_lookup
REFERENCING OLD TABLE AS oldtab FOR EACH STATEMENT EXECUTE FUNCTION update_codeintel_scip_documents_dereference_logs_delete();
