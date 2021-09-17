BEGIN;

ALTER TABLE lsif_data_definitions ADD COLUMN kind TEXT DEFAULT 'export' NOT NULL;
ALTER TABLE lsif_data_references ADD COLUMN kind TEXT DEFAULT 'import' NOT NULL;

ALTER TABLE lsif_data_definitions DROP CONSTRAINT lsif_data_definitions_pkey;
ALTER TABLE lsif_data_references DROP CONSTRAINT lsif_data_references_pkey;

ALTER TABLE ONLY lsif_data_definitions ADD CONSTRAINT lsif_data_definitions_pkey PRIMARY KEY (dump_id, kind, scheme, identifier);
ALTER TABLE ONLY lsif_data_references ADD CONSTRAINT lsif_data_references_pkey PRIMARY KEY (dump_id, kind, scheme, identifier);

COMMENT ON COLUMN lsif_data_definitions.kind IS 'The moniker kind.';
COMMENT ON COLUMN lsif_data_references.kind IS 'The moniker kind.';

COMMIT;
