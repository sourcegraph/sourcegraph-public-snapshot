-- +++
-- parent: 1528395873
-- +++

BEGIN;

-- Allow for lookup from upload id to commits that the upload can resolve queries for.
CREATE INDEX lsif_nearest_uploads_uploads ON lsif_nearest_uploads USING GIN(uploads);

-- Allow for lookup from commit to the set of commits that have analogous nearest uploads.
CREATE INDEX lsif_nearest_uploads_links_repository_id_ancestor_commit_bytea ON lsif_nearest_uploads_links(repository_id, ancestor_commit_bytea);

COMMIT;
