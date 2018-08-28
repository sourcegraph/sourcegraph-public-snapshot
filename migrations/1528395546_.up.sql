CREATE TABLE registry_extension_releases (
       id bigserial NOT NULL PRIMARY KEY,
       registry_extension_id integer NOT NULL REFERENCES registry_extensions(id) ON DELETE CASCADE ON UPDATE CASCADE,
       creator_user_id integer NOT NULL REFERENCES users(id),
       release_version citext,
       release_tag citext NOT NULL,
       manifest text NOT NULL,
       bundle text,
       created_at timestamp with time zone NOT NULL DEFAULT now(),
       deleted_at timestamp with time zone
);

CREATE UNIQUE INDEX registry_extension_releases_version ON registry_extension_releases(registry_extension_id, release_version) WHERE release_version IS NOT NULL;
CREATE INDEX registry_extension_releases_registry_extension_id ON registry_extension_releases(registry_extension_id, release_tag, created_at DESC) WHERE deleted_at IS NULL;

-- Transfer manifests from registry_extensions to registry_extension_releases. Use the 1st user's
-- UID as the creator if none other is available (as of this migration, extensions are experimental,
-- so this is acceptable).
INSERT INTO registry_extension_releases(registry_extension_id, creator_user_id, release_tag, manifest, created_at)
  SELECT id AS registry_extension_id, COALESCE(publisher_user_id, (SELECT id FROM users ORDER BY id ASC LIMIT 1)) AS creator_user_id,
    'release'::text as release_tag, manifest, updated_at AS created_at
  FROM registry_extensions
  WHERE manifest IS NOT NULL AND deleted_at IS NULL;

-- We don't need to retain the registry_extensions.manifest column for backcompat because as of this
-- migrations, extensions are experimental. But keep it anyway for now because otherwise we'd fail
-- the PostgreSQL backcompat step in CI.
--
-- ALTER TABLE registry_extensions DROP COLUMN manifest;
