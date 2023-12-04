DROP TABLE IF EXISTS codeintel_path_rank_inputs;

ALTER TABLE
    codeintel_path_ranks
ALTER COLUMN
    payload TYPE text USING payload :: text;

DROP AGGREGATE IF EXISTS sg_jsonb_concat_agg(jsonb);
