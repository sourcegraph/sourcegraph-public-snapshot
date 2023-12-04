ALTER TABLE codeintel_scip_metadata ADD COLUMN IF NOT EXISTS text_document_encoding text NOT NULL;
ALTER TABLE codeintel_scip_metadata ADD COLUMN IF NOT EXISTS protocol_version integer NOT NULL;

COMMENT ON COLUMN codeintel_scip_metadata.text_document_encoding IS 'The encoding of the text documents within this index. May affect range boundaries.';
COMMENT ON COLUMN codeintel_scip_metadata.protocol_version IS 'The version of the SCIP protocol used to encode this index.';
