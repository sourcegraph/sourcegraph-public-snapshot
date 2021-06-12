BEGIN;

CREATE INDEX lsif_uploads_associated_index_id ON lsif_uploads(associated_index_id);

COMMIT;
