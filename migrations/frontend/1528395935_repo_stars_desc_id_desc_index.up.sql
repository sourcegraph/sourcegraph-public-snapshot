-- +++
-- parent: 1528395934
-- +++

CREATE INDEX CONCURRENTLY IF NOT EXISTS repo_stars_desc_id_desc_idx
    ON repo USING btree (stars DESC NULLS LAST, id DESC) WHERE deleted_at IS NULL AND blocked IS NULL;
