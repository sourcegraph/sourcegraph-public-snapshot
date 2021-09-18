BEGIN;

-- Add nullable file_path column, for mapping documentationResult ID -> file_path
ALTER TABLE lsif_data_documentation_mappings ADD COLUMN file_path text;
COMMENT ON COLUMN lsif_data_documentation_mappings.file_path IS 'The document file path for the documentationResult, if any. e.g. the path to the file where the symbol described by this documentationResult is located, if it is a symbol.';

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE codeintel_schema_migrations SET dirty = 'f'
COMMIT;
