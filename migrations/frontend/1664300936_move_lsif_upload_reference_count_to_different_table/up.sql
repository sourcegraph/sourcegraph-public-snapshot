CREATE TABLE IF NOT EXISTS lsif_uploads_reference_counts (
    upload_id INTEGER UNIQUE NOT NULL,
    reference_count INTEGER NOT NULL
);

COMMENT ON TABLE lsif_uploads_reference_counts IS 'A less hot-path reference count for upload records.';

COMMENT ON COLUMN lsif_uploads_reference_counts.upload_id IS 'The identifier of the referenced upload.';

COMMENT ON COLUMN lsif_uploads_reference_counts.reference_count IS 'The number of references to the associated upload from other records (via lsif_references).';

ALTER TABLE
    lsif_uploads_reference_counts DROP CONSTRAINT IF EXISTS lsif_data_docs_search_private_repo_name_id_fk;

ALTER TABLE
    lsif_uploads_reference_counts
ADD
    CONSTRAINT lsif_data_docs_search_private_repo_name_id_fk FOREIGN KEY (upload_id) REFERENCES lsif_uploads(id) ON DELETE CASCADE;
