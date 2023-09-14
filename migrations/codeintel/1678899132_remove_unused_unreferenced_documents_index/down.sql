-- Undo the changes made in the up migration

CREATE INDEX IF NOT EXISTS codeintel_scip_documents_dereference_logs_last_removal_time_doc
ON codeintel_scip_documents_dereference_logs (last_removal_time, document_id);
