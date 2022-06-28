# Migration guides

- [Migrating from Oracle OpenGrok to Sourcegraph for code search](opengrok.md)
- [Back up or migrate to a new Sourcegraph instance](../deploy/migrate-backup.md)
- [Update notes](../updates/index.md)
- [Migrating from Sourcegraph 3.30.0, 3.30.1, and 3.30.2](3_30.md)
- [Migrating to Sourcegraph 3.31.x](3_31.md)

## Out of band migrations

Sourcegraph runs data migrations in the background while the instance is active instead of requiring a blocking migration
during startup or manual migrations requiring downtime.

Migrations are introduced at a particular version with an expected lifetime (a course of several versions). At the end of this
lifetime, the migration will be marked as deprecated and the instance will no longer be able to read the old data. This requires
that migrations finish prior to an upgrade to a version that no longer understands your instance's data.

The `Site Admin > Maintenance > Migrations` page shows the progress of all active migrations. This page will also display a 
prominent warning if when upgrade (or downgrade) would result in an instance that refuses to start due to an illegal migration
state.

![Unfinished migration warning](https://storage.googleapis.com/sourcegraph-assets/oobmigration-warning.png)

In this situation, upgrading to the next version will not result in any data loss, but all new instances will detect the illegal
migration state and refuse to start up with a fatal message (`Unfinished migrations`).

See [How to troubleshoot an unfinished migration](../how-to/unfinished_migration.md) for more information.

## Legacy guides

- [Migrating from Sourcegraph 2.13 to 3.0.0](3_0.md)
- [Migrating from Sourcegraph 3.x to 3.7.2+](3_7.md)
- [Migrating from Sourcegraph 3.x to 3.11](3_11.md)
