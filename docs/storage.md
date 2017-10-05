# PostgreSQL

Sourcegraph stores most data in a
[PostgreSQL database](http://www.postgresql.org). Git repositories,
uploaded user content (e.g., image attachments in issues) are stored
on the filesystem.

# Initializing PostgreSQL

After installing PostgreSQL, set up up a `sourcegraph` user and database:

```
sudo su - postgres # this line only needed for Linux
createdb
createuser --superuser sourcegraph
psql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';"
createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph
```

Then update your `postgresql.conf` default timezone to UTC. Determine the location
of your `postgresql.conf` by running `psql -c 'show config_file;'`. Update the line beginning
with `timezone =` to the following:

```
timezone = 'UTC'
```

Finally, restart your database server (mac: `brew services restart postgresql`)

# Configuring PostgreSQL

The Sourcegraph server reads PostgreSQL connection configuration from
the
[`PG*` environment variables](http://www.postgresql.org/docs/current/static/libpq-envars.html);
for example:

```
PGPORT=5432
PGUSER=sourcegraph
PGPASSWORD=sourcegraph
PGDATABASE=sourcegraph
PGSSLMODE=disable
```

To test the environment's credentials, run `psql` (the PostgreSQL CLI
client) with the `PG*` environment variables set. If you see a
database prompt, then the environment's credentials are valid.

# Migrations

Documented in [../migrations/README.md](../migrations/README.md)

# Style guide

Here is the preferred style going forward. Existing tables may be inconsistent with this style.

## Recommended columns for all tables

- `id` auto increment primary key.
- `created_at` not null default `now()` set when a row is first inserted and never updated after that.
- `updated_at` not null default `now()` set when a row is first inserted and updated on every update.
- `deleted_at` set to a not null timestamp to indicate the row is deleted (called soft deleting). This is preferred over hard deleting data from our db (see discussion section below).
  - When querying the db, rows with a non-null `deleted_at` should be excluded.

The timestamps are useful for forensics if something goes wrong, they do not necessarily need to be used or exposed by our graphql APIs. There is no harm in exposing them though.

Example:

```sql
CREATE TABLE "widgets" (
	"id" bigserial NOT NULL PRIMARY KEY,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	"updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	"deleted_at" TIMESTAMP WITH TIME ZONE,
);
```

## Hard vs soft deletes

Definitions:
- A "hard" delete is when rows are deleted using `DELETE FROM table WHERE ...`
- A "soft" delete is when rows are deleted using `UPDATE table SET deleted_at = now() WHERE ...`

Hard deletes are hard to recover from if something goes wrong (application bug, bad migration, manual query, etc.). This usually involves restoring from a backup and it is hard to target only the data affected by the bad delete.

Soft deletes are easier to recover from once you determine what happened. You can simply find the affected rows and `UPDATE table SET deleted_at = null WHERE ...`.

### Dealing with unique constraints

Soft deleting data has implications for unique constraints.

Consider a hypothetical schema:
```sql
CREATE TABLE "orgs" (
	"id" serial NOT NULL PRIMARY KEY
);

CREATE TABLE "users" (
	"id" serial NOT NULL PRIMARY KEY
);

CREATE TABLE "users_orgs" (
	"id" serial NOT NULL PRIMARY KEY,
	"user_id" integer NOT NULL,
	"org_id" integer NOT NULL,

	CONSTRAINT user_orgs_references_orgs
	FOREIGN KEY (org_id)
	REFERENCES orgs (id) ON DELETE RESTRICT,

	CONSTRAINT users_references_users
	FOREIGN KEY (user_id)
	REFERENCES users (id) ON DELETE RESTRICT,

	UNIQUE (user_id, org_id)
);
```

#### Hard delete case

Removing a user from an org deletes the row from `user_orgs`.

Adding a user inserts a row to `user_orgs`. If the user is already a user of the org, the insert fails.

If we wanted to keep a record of membership, it would need to be in a separate audit log table.

#### Soft delete case

Removing a user from an org sets a non-null timestamp on the `deleted_at` column for the row.

Adding a user to an org sets `deleted_at = null` if there is already an existing record for that combination of `user_id` and `org_id`, else a new record is inserted.

Alternatively, we could remove the unique constraint on `user_id` and `org_id` and always insert in the add user case (after checking to see if the user is in the org). This would then function as an audit log table.

The decision here can be made on a table by table basis.

## Use foreign keys

If you have a column that references another column in the database, add a foreign key constraint.

Foreign key constraints should not cascade deletes. We don't want to accidentally delete a lot of data.

There are reasons to not use foreign keys at scale, but we are not at scale and we can drop these in the future if they become a problem.

## Table names

Tables are plural (e.g. repositories, users, comments, etc.).

Join tables should be named based on the two tables being joined (e.g. `foo_bar` joins `foo` and `bar`).

## Validation

To the extent that certain fields require validation (e.g. username) we should perform that validation in client AND EITHER the database when possible, OR the graphql api. This results in the best experience for the client, and protects us from corrupt data.

## Trigger functions

Trigger functions perform some action when data is inserted or updated. Don't use trigger functions.
