-- We didn't actually drop this column (to avoid failing the PostgreSQL CI step). A future migration
-- should drop it.
--
-- ALTER TABLE registry_extensions ADD COLUMN manifest text;

-- Transfer the latest 'release' manifest back to the registry_extensions.manifest column.
UPDATE registry_extensions x
  SET manifest=r.manifest, updated_at=r.created_at
  FROM registry_extension_releases r
  WHERE x.id=r.registry_extension_id AND r.release_tag='release' AND r.deleted_at IS NULL AND
  NOT EXISTS (
    SELECT 1 FROM registry_extension_releases r2
    WHERE r2.registry_extension_id=r.registry_extension_id AND r2.release_tag=r.release_tag AND
          r2.created_at > r.created_at
  );

DROP TABLE registry_extension_releases;
