BEGIN;

-- faster to supply default than manual update
ALTER TABLE lsif_data_references ADD COLUMN schema_version int DEFAULT 1 NOT NULL;
ALTER TABLE lsif_data_references ADD COLUMN num_locations int DEFAULT 0 NOT NULL;

-- drop default after all existing columns have been set
ALTER TABLE lsif_data_references ALTER COLUMN schema_version DROP DEFAULT;
ALTER TABLE lsif_data_references ALTER COLUMN num_locations DROP DEFAULT;

COMMENT ON COLUMN lsif_data_references.schema_version IS 'The schema version of this row - used to determine presence and encoding of data.';
COMMENT ON COLUMN lsif_data_references.num_locations IS 'The number of locations stored in the data field.';

CREATE TABLE lsif_data_references_schema_versions (
    dump_id integer NOT NULL,
    min_schema_version integer,
    max_schema_version integer
);
ALTER TABLE lsif_data_references_schema_versions ADD PRIMARY KEY (dump_id);

COMMENT ON TABLE lsif_data_references_schema_versions IS 'Tracks the range of schema_versions for each upload in the lsif_data_references table.';
COMMENT ON COLUMN lsif_data_references_schema_versions.dump_id IS 'The identifier of the associated dump in the lsif_uploads table.';
COMMENT ON COLUMN lsif_data_references_schema_versions.min_schema_version IS 'A lower-bound on the `lsif_data_references.schema_version` where `lsif_data_references.dump_id = dump_id`.';
COMMENT ON COLUMN lsif_data_references_schema_versions.max_schema_version IS 'An upper-bound on the `lsif_data_references.schema_version` where `lsif_data_references.dump_id = dump_id`.';

-- Ensure that there is a lsif_data_references_schema_versions record for each distinct dump_id in
-- the lsif_data_references table. All reference versions in the database at this point necessarily
-- have a schema_version of 1, so we can gather all of the dump ids quickly by scanning the metadata
-- table. Grouping the references table directly would take too much time in a migration.

INSERT INTO lsif_data_references_schema_versions
    SELECT
        dump_id,
        1 AS min_schema_version,
        1 AS max_schema_version
    FROM
        lsif_data_metadata
    ORDER BY
        lsif_data_metadata;

-- On every insert into lsif_data_references, we need to make sure we have an associated row in the
-- lsif_data_references_schema_versions table. We do not currently care about cleaning the table up
-- (we will do this asynchronously).
--
-- We use FOR EACH STATEMENT here because we batch insert into this table. Running the trigger per
-- statement rather than per row will save a ton of extra work. Running over batch inserts lets us
-- do a GROUP BY on the new table and effectively upsert our new ranges.
--
-- Note that the only places where data is _modified_ in this database is during migrations, which
-- will necessarily update this table's bounds for any migrated index records.

CREATE OR REPLACE FUNCTION update_lsif_data_references_schema_versions_insert() RETURNS trigger AS $$ BEGIN
    INSERT INTO
        lsif_data_references_schema_versions
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
        min_schema_version = LEAST(lsif_data_references_schema_versions.min_schema_version, EXCLUDED.min_schema_version),
        max_schema_version = GREATEST(lsif_data_references_schema_versions.max_schema_version, EXCLUDED.max_schema_version);

    RETURN NULL;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER lsif_data_references_schema_versions_insert
AFTER INSERT ON lsif_data_references REFERENCING NEW TABLE AS newtab
FOR EACH STATEMENT EXECUTE PROCEDURE update_lsif_data_references_schema_versions_insert();

COMMIT;
