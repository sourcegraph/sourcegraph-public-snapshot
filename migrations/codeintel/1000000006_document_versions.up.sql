BEGIN;

CREATE TABLE lsif_data_documents_schema_versions (
    dump_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer
);
ALTER TABLE lsif_data_documents_schema_versions ADD PRIMARY KEY (dump_id);

COMMENT ON TABLE lsif_data_documents_schema_versions IS 'Tracks the range of schema_versions for each upload in the lsif_data_documents table.';
COMMENT ON COLUMN lsif_data_documents_schema_versions.dump_id IS 'The identifier of the associated dump in the lsif_uploads table.';
COMMENT ON COLUMN lsif_data_documents_schema_versions.min_schema_version IS 'A lower-bound on the `lsif_data_documents.schema_version` where `lsif_data_documents.dump_id = dump_id`.';
COMMENT ON COLUMN lsif_data_documents_schema_versions.max_schema_version IS 'An upper-bound on the `lsif_data_documents.schema_version` where `lsif_data_documents.dump_id = dump_id`.';

-- Ensure that there is a lsif_data_documents_schema_versions record for each distinct dump_id in
-- the lsif_data_documents table. We are denormalizing schema counts here, but unfortunately we
-- have already started the migration and don't have a known schema version to put here. This will
-- preclude us from doing a distinct on dump_id, or reading from the smaller metadata table.
--
-- We'll group these now. This query takes around 8 seconds on the Cloud database, so this should
-- not cause a problem in any known instance.
INSERT INTO lsif_data_documents_schema_versions
    SELECT
        dump_id,
        min(schema_version) as min_schema_version,
        max(schema_version) as max_schema_version
    FROM
        lsif_data_documents
    GROUP BY
        dump_id
    ORDER BY
        dump_id;

-- On every insert into lsif_data_documents, we need to make sure we have an associated row in the
-- lsif_data_documents_schema_versions table. We do not currently care about cleaning the table up
-- (we will do this asynchronously).
--
-- We use FOR EACH STATEMENT here because we batch insert into this table. Running the trigger per
-- statement rather than per row will save a ton of extra work. Running over batch inserts lets us
-- do a GROUP BY on the new table and effectively upsert our new ranges.
--
-- Note that the only places where data is _modified_ in this database is during migrations, which
-- will necessarily update this table's bounds for any migrated index records.

CREATE OR REPLACE FUNCTION update_lsif_data_documents_schema_versions_insert() RETURNS trigger AS $$ BEGIN
    INSERT INTO
        lsif_data_documents_schema_versions
    SELECT
        dump_id,
        MIN(schema_version) as min_schema_version,
        MAX(schema_version) as max_schema_version
    FROM
        newtab
    GROUP BY
        dump_id
    ON CONFLICT (dump_id) DO UPDATE SET
        -- Update with min(old_min, new_min) and max(old_max, new_max)
        min_schema_version = LEAST(lsif_data_documents_schema_versions.min_schema_version, EXCLUDED.min_schema_version),
        max_schema_version = GREATEST(lsif_data_documents_schema_versions.max_schema_version, EXCLUDED.max_schema_version);

    RETURN NULL;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER lsif_data_documents_schema_versions_insert
AFTER INSERT ON lsif_data_documents REFERENCING NEW TABLE AS newtab
FOR EACH STATEMENT EXECUTE PROCEDURE update_lsif_data_documents_schema_versions_insert();

COMMIT;
