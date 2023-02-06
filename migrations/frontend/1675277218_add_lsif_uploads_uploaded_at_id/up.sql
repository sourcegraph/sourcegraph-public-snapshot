CREATE INDEX CONCURRENTLY IF NOT EXISTS lsif_uploads_uploaded_at_id ON lsif_uploads(uploaded_at desc, id) WHERE state != 'deleted';
