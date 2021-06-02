BEGIN;

ALTER TABLE out_of_band_migrations ADD COLUMN introduced text;
UPDATE out_of_band_migrations SET introduced = concat(introduced_version_major, '.', introduced_version_minor);
ALTER TABLE out_of_band_migrations ALTER COLUMN introduced SET NOT NULL;
ALTER TABLE out_of_band_migrations DROP COLUMN introduced_version_major;
ALTER TABLE out_of_band_migrations DROP COLUMN introduced_version_minor;

ALTER TABLE out_of_band_migrations ADD COLUMN deprecated text;
UPDATE out_of_band_migrations SET deprecated = concat(deprecated_version_major, '.', deprecated_version_minor) WHERE deprecated_version_major IS NOT NULL;
ALTER TABLE out_of_band_migrations DROP COLUMN deprecated_version_major;
ALTER TABLE out_of_band_migrations DROP COLUMN deprecated_version_minor;

COMMIT;
