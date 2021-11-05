CREATE INDEX CONCURRENTLY IF NOT EXISTS repo_stars_desc_id_asc_idx
    ON repo USING btree (stars DESC NULLS LAST, id ASC) WHERE deleted_at IS NULL;
