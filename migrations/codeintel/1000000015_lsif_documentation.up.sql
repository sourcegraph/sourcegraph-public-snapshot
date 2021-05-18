BEGIN;

CREATE TABLE IF NOT EXISTS lsif_data_documentation_pages (
    dump_id integer NOT NULL,
    path_id TEXT,
    data bytea
);

ALTER TABLE lsif_data_documentation_pages ADD PRIMARY KEY (dump_id, path_id);

COMMENT ON TABLE lsif_data_documentation_pages IS 'Associates documentation pathIDs to their documentation page hierarchy chunk.';
COMMENT ON COLUMN lsif_data_documentation_pages.dump_id IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';
COMMENT ON COLUMN lsif_data_documentation_pages.path_id IS 'The documentation page path ID, see see GraphQL codeintel.schema:documentationPage for what this is.';
COMMENT ON COLUMN lsif_data_documentation_pages.data IS 'A gob-encoded payload conforming to a `type DocumentationPageData struct` pointer (lib/codeintel/semantic/types.go)';

COMMIT;
