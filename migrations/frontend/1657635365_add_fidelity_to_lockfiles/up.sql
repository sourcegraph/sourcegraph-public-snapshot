ALTER TABLE codeintel_lockfiles
  ADD COLUMN IF NOT EXISTS fidelity text NOT NULL DEFAULT 'flat';

COMMENT ON COLUMN codeintel_lockfiles.fidelity IS 'Fidelity of the dependency graph thats persisted, whether it is a flat list, a whole graph, circular graph, ...';
