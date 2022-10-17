ALTER TABLE codeintel_lockfiles
  ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT NOW(),
  ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT NOW()
  ;

COMMENT ON COLUMN codeintel_lockfiles.created_at IS 'Time when lockfile was indexed';
COMMENT ON COLUMN codeintel_lockfiles.updated_at IS 'Time when lockfile index was updated';
