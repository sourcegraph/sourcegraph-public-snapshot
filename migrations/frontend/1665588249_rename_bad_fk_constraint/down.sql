-- Put the weird name back
ALTER TABLE
    lsif_uploads_reference_counts DROP CONSTRAINT IF EXISTS lsif_uploads_reference_counts_upload_id_fk;

ALTER TABLE
    lsif_uploads_reference_counts DROP CONSTRAINT IF EXISTS lsif_data_docs_search_private_repo_name_id_fk;

ALTER TABLE
    lsif_uploads_reference_counts
ADD
    CONSTRAINT lsif_data_docs_search_private_repo_name_id_fk FOREIGN KEY (upload_id) REFERENCES lsif_uploads(id) ON DELETE CASCADE;
