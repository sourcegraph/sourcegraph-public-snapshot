BEGIN;

CREATE TABLE user_permissions (
  user_id     integer NOT NULL REFERENCES users (id),
  permission  text NOT NULL,
  object_type text NOT NULL,
  object_ids  bytea NOT NULL,
  updated_at  timestamptz NOT NULL
);

ALTER TABLE user_permissions
ADD CONSTRAINT user_permissions_perm_object_unique
UNIQUE (user_id, permission, object_type);

COMMIT;
