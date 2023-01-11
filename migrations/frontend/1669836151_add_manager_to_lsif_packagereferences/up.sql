ALTER TABLE lsif_packages ADD COLUMN IF NOT EXISTS manager TEXT NOT NULL DEFAULT '';
ALTER TABLE lsif_references ADD COLUMN IF NOT EXISTS manager TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN lsif_packages.manager IS 'The package manager name.';
COMMENT ON COLUMN lsif_references.manager IS 'The package manager name.';
