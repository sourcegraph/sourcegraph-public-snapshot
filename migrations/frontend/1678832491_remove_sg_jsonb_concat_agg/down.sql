-- Undo the changes made in the up migration

CREATE OR REPLACE AGGREGATE sg_jsonb_concat_agg(jsonb) (
    SFUNC = jsonb_concat,
    STYPE = jsonb,
    INITCOND = '{}'
);

