-- Fix registry_extensions so that extensions can't have the same publisher and name.

-- Delete extensions that have the same publisher and the same name. It was never intended for such
-- conflicting extensions to be created. At the time of this migration, extensions were still
-- experimental, so this will not affect any customers.
DELETE FROM registry_extensions x1
USING registry_extensions x2
WHERE
  x1.name = x2.name AND (
    (x1.publisher_user_id IS NOT NULL AND x1.publisher_user_id = x2.publisher_user_id) OR
    (x1.publisher_org_id IS NOT NULL AND x1.publisher_org_id = x2.publisher_org_id)
  ) AND x1.id != x2.id;

DROP INDEX registry_extensions_publisher_name;
CREATE UNIQUE INDEX registry_extensions_publisher_name ON registry_extensions(COALESCE(publisher_user_id, 0), COALESCE(publisher_org_id, 0), name) WHERE deleted_at IS NULL;
