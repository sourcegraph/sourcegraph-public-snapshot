-- +++
-- parent: 1000000019
-- +++

BEGIN;

CREATE TYPE presentation_type_enum AS ENUM ('LINE', 'PIE');
ALTER TABLE insight_view
    ADD COLUMN IF NOT EXISTS presentation_type presentation_type_enum NOT NULL DEFAULT 'LINE',
    ALTER COLUMN other_threshold type FLOAT8; -- Changing this because the GraphQL float type is a float64.

COMMENT ON COLUMN insight_view.presentation_type IS 'The basic presentation type for the insight view. (e.g Line, Pie, etc.)';

COMMIT;
