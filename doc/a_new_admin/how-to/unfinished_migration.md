# How to troubleshoot an unfinished migration

This document will take you through how to resolve an 'unfinished migration' error. Between two versions of Sourcegraph an out-of-band migration routine may be deprecated. In this new version, the instance will only know how to read the "new" format of data created by the migration. If your instance did not complete the migration, then there will be data left in the "old" format that will no longer be readable. Instead of allowing the instance to run in this unpredictable state, it dies with an 'unfinished migration' error on startup.

The error will look something like this:

```
ERROR: Unfinished migrations. Please revert Sourcegraph to the previous version and wait for the following migrations to complete.

- migration 12 expected to be at 100% (at 63.27%)
- migration 16 expected to be at 100% (at 41.84%)
```

## Prerequisites

* This document assumes that you were attempting an upgrade when an error occurred.

## Steps to resolve

1. Roll back from v3.X to the previous working version v3.Y
2. Navigate to the `Site Admin > Maintenance > Migrations` page and note the set of unfinished migrations (progress < 100%) deprecated after v3.X but not after v3.Y
  - These are the set of migrations that need to be completed before upgrading back to 3.X
3. If a migration is making upwards progress, simply wait for it to complete
  - Note that the speed of some migrations may be tunable via environment variables or configuration
4. If a migration has stalled and is no longer making progress, check the recent errors associated with migration in the UI
  - Resolving these errors should unclog the migration
  - If there are no errors then the migration is broken - contact the engineering team

Each migration notes the engineering team that is able to resolve associated errors.

As an escape hatch, the environment variable `SRC_DISABLE_OOBMIGRATION_VALIDATION` can be set to `true` on the `frontend` and `worker` services to disable the startup check.

## Further resources

* [Out of band migrations](../migration/index.md)
* [Sourcegraph - Upgrading Sourcegraph to a new version](https://docs.sourcegraph.com/admin/updates)
