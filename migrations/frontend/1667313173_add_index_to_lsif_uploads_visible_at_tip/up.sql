CREATE INDEX IF NOT EXISTS lsif_uploads_visible_at_tip_is_default_branch ON lsif_uploads_visible_at_tip(upload_id)
WHERE
    is_default_branch;
