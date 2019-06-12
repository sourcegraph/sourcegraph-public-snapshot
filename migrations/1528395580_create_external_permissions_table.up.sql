BEGIN;

CREATE TABLE external_permissions (
  account_id  integer NOT NULL REFERENCES user_external_accounts (id),
  permission  text NOT NULL,
  object_type text NOT NULL,
  object_ids  bytea NOT NULL,
  updated_at  timestamptz NOT NULL,
  expired_at  timestamptz NOT NULL
);

ALTER TABLE external_permissions
ADD CONSTRAINT external_permissions_account_perm_object_unique
UNIQUE (account_id, permission, object_type);

COMMIT;
