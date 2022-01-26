-- +++
-- parent: 1000000016
-- +++

BEGIN;

CREATE TABLE IF NOT EXISTS lsif_data_documentation_path_info (
    dump_id integer NOT NULL,
    path_id TEXT,
    data bytea
);

ALTER TABLE lsif_data_documentation_path_info ADD PRIMARY KEY (dump_id, path_id);

COMMENT ON TABLE lsif_data_documentation_path_info IS 'Associates documentation page pathIDs to information about what is at that pathID, its immediate children, etc.';
COMMENT ON COLUMN lsif_data_documentation_path_info.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';
COMMENT ON COLUMN lsif_data_documentation_path_info.path_id IS 'The documentation page path ID, see see GraphQL codeintel.schema:documentationPage for what this is.';
COMMENT ON COLUMN lsif_data_documentation_path_info.data IS 'A gob-encoded payload conforming to a `type DocumentationPathInoData struct` pointer (lib/codeintel/semantic/types.go)';

COMMIT;
