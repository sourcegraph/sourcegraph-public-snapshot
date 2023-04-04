CREATE OR REPLACE FUNCTION update_codeintel_path_ranks_statistics_columns() RETURNS trigger
LANGUAGE plpgsql
AS $$ BEGIN
    SELECT
        COUNT(r.v) AS num_paths,
        SUM(LOG(2, r.v::int + 1)) AS refcount_logsum
    INTO
        NEW.num_paths,
        NEW.refcount_logsum
    FROM jsonb_each(
        CASE WHEN NEW.payload::text = 'null'
            THEN '{}'::jsonb
            ELSE COALESCE(NEW.payload, '{}'::jsonb)
        END
    ) r(k, v);

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
