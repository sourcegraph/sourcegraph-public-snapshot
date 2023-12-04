ALTER TABLE
    lsif_dirty_repositories
ADD
    COLUMN IF NOT EXISTS set_dirty_at timestamptz NOT NULL DEFAULT NOW();
