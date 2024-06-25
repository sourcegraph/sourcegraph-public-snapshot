-- CREATE TYPE IF NOT EXISTS is not a thing, so we must drop it here,
DROP TYPE IF EXISTS query_version_enum;
CREATE TYPE query_version_enum AS ENUM ('V1', 'V2', 'V3', 'V4');

ALTER TABLE notebooks ADD COLUMN IF NOT EXISTS query_version query_version_enum NOT NULL DEFAULT 'V3';
