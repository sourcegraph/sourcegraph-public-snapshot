BEGIN;

ALTER TABLE IF EXISTS repo ADD COLUMN blocked jsonb DEFAULT NULL;

CREATE OR REPLACE FUNCTION repo_block(reason text, at timestamptz) RETURNS jsonb AS
$$
SELECT jsonb_build_object(
    'reason', reason,
    'at', extract(epoch from timezone('utc', at))::bigint
);
$$ LANGUAGE SQL STRICT IMMUTABLE LEAKPROOF;

CREATE INDEX repo_is_blocked_idx ON repo USING BTREE ((blocked IS NOT NULL));

COMMIT;
