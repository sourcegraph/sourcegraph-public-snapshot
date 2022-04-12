DO
$$
    DECLARE
        tsdb_ext PG_EXTENSION%ROWTYPE;
    BEGIN
        -- To ensure we only excute this once (in case of manual executions) we will check if the TimescaleDB extension exists, and only if so
        -- perform this table migration
        SELECT * FROM pg_extension WHERE extname = 'timescaledb' LIMIT 1 INTO tsdb_ext;
        IF NOT found THEN
        ELSE
            -- Perform a table swap - create a new table and rename the hypertable
            CREATE TABLE series_points_vanilla
            (
                LIKE series_points INCLUDING ALL
            );
            ALTER    TABLE series_points
                RENAME TO series_points_timescale;
            ALTER TABLE series_points_vanilla
                RENAME TO series_points;

            -- Copy all of the data and insert into the new table.
            INSERT INTO series_points (SELECT * FROM series_points_timescale);

            -- Drop the old hypertable (the extension will propagate and drop all of the hypertable stuff)
            DROP TABLE series_points_timescale CASCADE;

            -- Last, remove the extension
            DROP EXTENSION IF EXISTS timescaledb;
        END IF;
    END
$$;
