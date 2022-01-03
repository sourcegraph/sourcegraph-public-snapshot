# Migrations

For each of our Postgres instances, we define a sequence of SQL schema commands that must be applied before the database is in the state the application expects. We also support zero-downtime upgrades between one minor version in both Sourcegraph Cloud, managed instances, and enterprise instances.

In development environments, these migrations are applied automatically on application startup. This is a specific choice to keep the response latency small during development. In production environments, a typical upgrade requires that the site-administrator first run a `migrator` service to prepare the database schema for the new version of the application. This is a type of _database-first_ deployment (opposed to _code-first_ deployments), where database migrations are applied prior to the corresponding code change.

Database migrations may be applied arbitrarily long before the new version is deployed. This implies that an old version of Sourcegraph (up to one minor version) can run against a new schema. This requires that all of our database schema changes be *backwards-compatible* with respect to the previous release; any changes to the database schema that would alter the behavior of an old instance is disallowed (and enforced in CI).

Some migrations are difficult to do in a single step. For instance, renaming a column, table, or view, or adding a column with a non-nullable constraint will all break existing code that accesses that table or view. In order to do such changes you may need to break your changes into several parts separated by a minor release.

For example, a non-nullable column can be added to an existing table with the following steps:

- Add a nullable column to the table
- Update the code to always populate this row on writes
- Deploy to Sourcegraph.com
- Add a non-nullable constraint to the table
- Deploy to Sourcegraph.com
