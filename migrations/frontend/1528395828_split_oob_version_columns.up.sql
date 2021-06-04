BEGIN;

--
-- Switch out introduced column

ALTER TABLE out_of_band_migrations ADD COLUMN introduced_version_major int;
ALTER TABLE out_of_band_migrations ADD COLUMN introduced_version_minor int;

WITH t(id, parts) AS (
    SELECT
        id,
        regexp_matches(introduced, E'^(\\d+)\.(\\d+)')
    FROM
        out_of_band_migrations
)
UPDATE out_of_band_migrations SET
    introduced_version_major = parts[1]::int,
    introduced_version_minor = parts[2]::int
FROM t WHERE t.id = out_of_band_migrations.id;

ALTER TABLE out_of_band_migrations ALTER COLUMN introduced_version_major SET NOT NULL;
ALTER TABLE out_of_band_migrations ALTER COLUMN introduced_version_minor SET NOT NULL;
ALTER TABLE out_of_band_migrations DROP COLUMN introduced;

--
-- Switch out deprecation column (keep nullable, no data exists yet)

ALTER TABLE out_of_band_migrations ADD COLUMN deprecated_version_major int;
ALTER TABLE out_of_band_migrations ADD COLUMN deprecated_version_minor int;
ALTER TABLE out_of_band_migrations DROP COLUMN deprecated;

COMMENT ON COLUMN out_of_band_migrations.introduced_version_major IS 'The Sourcegraph version (major component) in which this migration was first introduced.';
COMMENT ON COLUMN out_of_band_migrations.introduced_version_minor IS 'The Sourcegraph version (minor component) in which this migration was first introduced.';
COMMENT ON COLUMN out_of_band_migrations.deprecated_version_major IS 'The lowest Sourcegraph version (major component) that assumes the migration has completed.';
COMMENT ON COLUMN out_of_band_migrations.deprecated_version_minor IS 'The lowest Sourcegraph version (minor component) that assumes the migration has completed.';

COMMIT;
