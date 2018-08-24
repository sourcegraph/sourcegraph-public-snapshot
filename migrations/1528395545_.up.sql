-- The names table enforces that users and organizations are a single namespace. If a user named
-- "alice" exists, no organization named "alice" may be created, and vice versa.
--
-- In the past, it was possible to create a user and organization with the same name. This migration
-- must not fail if such a conflict exists (because failing migrations are a huge upgrade pain). The
-- goal is to prevent new conflicts from being created. In a future release, when admins have been
-- given enough notice, we can enforce this retroactively, too.
CREATE TABLE names (
       name citext NOT NULL PRIMARY KEY,
       user_id integer REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
       org_id integer REFERENCES orgs(id) ON DELETE CASCADE ON UPDATE CASCADE,
       CHECK (user_id IS NOT NULL OR org_id IS NOT NULL)
);

INSERT INTO names(name, user_id) SELECT username AS name, id AS user_id FROM users WHERE deleted_at IS NULL;

-- If any organization names conflict with usernames, allow the conflicting organization to
-- remain. The UI will handle notifying admins of these conflicts; we don't want this migration to
-- fail. Creation of new users or organizations with conflicting names is prevented.
INSERT INTO names(name, org_id) SELECT name, id AS org_id FROM orgs WHERE name NOT IN (SELECT name FROM names) AND deleted_at IS NULL;
