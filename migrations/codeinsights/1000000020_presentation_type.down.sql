BEGIN;

ALTER TABLE insight_view
    ALTER COLUMN other_threshold TYPE FLOAT4,
    DROP COLUMN IF EXISTS presentation_type;

DROP TYPE presentation_type_enum;

COMMIT;
