-- +++
-- parent: 1000000029
-- +++

BEGIN;

CREATE TABLE lsif_data_implementations (
    dump_id        INTEGER NOT NULL,
    scheme         TEXT    NOT NULL,
    identifier     TEXT    NOT NULL,
    data           BYTEA           ,
    schema_version INTEGER NOT NULL,
    num_locations  INTEGER NOT NULL
);

COMMENT ON TABLE  lsif_data_implementations                IS 'Associates (document, range) pairs with the implementation monikers attached to the range.';
COMMENT ON COLUMN lsif_data_implementations.dump_id        IS 'The identifier of the associated dump in the lsif_uploads table (state=completed).';
COMMENT ON COLUMN lsif_data_implementations.scheme         IS 'The moniker scheme.';
COMMENT ON COLUMN lsif_data_implementations.identifier     IS 'The moniker identifier.';
COMMENT ON COLUMN lsif_data_implementations.data           IS 'A gob-encoded payload conforming to an array of [LocationData](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@3.26/-/blob/enterprise/lib/codeintel/semantic/types.go#L106:6) types.';
COMMENT ON COLUMN lsif_data_implementations.schema_version IS 'The schema version of this row - used to determine presence and encoding of data.';
COMMENT ON COLUMN lsif_data_implementations.num_locations  IS 'The number of locations stored in the data field.';

ALTER TABLE ONLY lsif_data_implementations ADD CONSTRAINT lsif_data_implementations_pkey PRIMARY KEY (dump_id, scheme, identifier);

CREATE INDEX lsif_data_implementations_dump_id_schema_version ON lsif_data_implementations (dump_id, schema_version);

CREATE TABLE lsif_data_implementations_schema_versions (
    dump_id            INTEGER NOT NULL,
    min_schema_version INTEGER         ,
    max_schema_version INTEGER
);

COMMENT ON TABLE lsif_data_implementations_schema_versions                     IS 'Tracks the range of schema_versions for each upload in the lsif_data_implementations table.';
COMMENT ON COLUMN lsif_data_implementations_schema_versions.dump_id            IS 'The identifier of the associated dump in the lsif_uploads table.';
COMMENT ON COLUMN lsif_data_implementations_schema_versions.min_schema_version IS 'A lower-bound on the `lsif_data_implementations.schema_version` where `lsif_data_implementations.dump_id = dump_id`.';
COMMENT ON COLUMN lsif_data_implementations_schema_versions.max_schema_version IS 'An upper-bound on the `lsif_data_implementations.schema_version` where `lsif_data_implementations.dump_id = dump_id`.';

ALTER TABLE ONLY lsif_data_implementations_schema_versions ADD CONSTRAINT lsif_data_implementations_schema_versions_pkey PRIMARY KEY (dump_id);

CREATE INDEX lsif_data_implementations_schema_versions_dump_id_schema_version_bounds ON lsif_data_implementations_schema_versions (dump_id, min_schema_version, max_schema_version);

-- On every insert into lsif_data_implementations, we need to make sure we have an associated row in the
-- lsif_data_implementations_schema_versions table. We do not currently care about cleaning the table up
-- (we will do this asynchronously).
--
-- We use FOR EACH STATEMENT here because we batch insert into this table. Running the trigger per
-- statement rather than per row will save a ton of extra work. Running over batch inserts lets us
-- do a GROUP BY on the new table and effectively upsert our new ranges.
--
-- Note that the only places where data is _modified_ in this database is during migrations, which
-- will necessarily update this table's bounds for any migrated index records.

CREATE OR REPLACE FUNCTION update_lsif_data_implementations_schema_versions_insert() RETURNS trigger AS $$ BEGIN
    INSERT INTO
        lsif_data_implementations_schema_versions
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
        min_schema_version = LEAST   (lsif_data_implementations_schema_versions.min_schema_version, EXCLUDED.min_schema_version),
        max_schema_version = GREATEST(lsif_data_implementations_schema_versions.max_schema_version, EXCLUDED.max_schema_version);

    RETURN NULL;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER lsif_data_implementations_schema_versions_insert
    AFTER INSERT ON lsif_data_implementations REFERENCING NEW TABLE AS newtab
    FOR EACH STATEMENT EXECUTE PROCEDURE update_lsif_data_implementations_schema_versions_insert();

COMMIT;
