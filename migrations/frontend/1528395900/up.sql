CREATE INDEX CONCURRENTLY IF NOT EXISTS lsif_uploads_repository_id_commit ON lsif_uploads(repository_id, commit);
