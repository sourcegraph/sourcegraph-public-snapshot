CREATE OR REPLACE FUNCTION update_codeintel_path_ranks_statistics_columns() RETURNS trigger
LANGUAGE plpgsql
AS $$ BEGIN
    SELECT
        COUNT(r.v) AS num_paths,
        MIN(r.v::int) AS min_reference_count,
        MAX(r.v::int) AS max_reference_count,
        SUM(r.v::int) AS sum_reference_count
    INTO
        NEW.num_paths,
        NEW.min_reference_count,
        NEW.max_reference_count,
        NEW.sum_reference_count
    FROM jsonb_each(NEW.payload) r(k, v);

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS insert_codeintel_path_ranks_statistics ON codeintel_path_ranks;
DROP TRIGGER IF EXISTS update_codeintel_path_ranks_statistics ON codeintel_path_ranks;

CREATE TRIGGER insert_codeintel_path_ranks_statistics
BEFORE INSERT ON codeintel_path_ranks
FOR EACH ROW
EXECUTE FUNCTION update_codeintel_path_ranks_statistics_columns();

CREATE TRIGGER update_codeintel_path_ranks_statistics
BEFORE UPDATE ON codeintel_path_ranks
FOR EACH ROW
WHEN ((new.* IS DISTINCT FROM old.*))
EXECUTE FUNCTION update_codeintel_path_ranks_statistics_columns();
