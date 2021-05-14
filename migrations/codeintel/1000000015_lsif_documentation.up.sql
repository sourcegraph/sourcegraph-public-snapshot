BEGIN;

CREATE TABLE IF NOT EXISTS lsif_data_documentation_pages (
    dump_id integer NOT NULL,
    path_id TEXT,
    data bytea
);

ALTER TABLE lsif_data_documentation_pages ADD PRIMARY KEY (dump_id, path_id);

COMMIT;
