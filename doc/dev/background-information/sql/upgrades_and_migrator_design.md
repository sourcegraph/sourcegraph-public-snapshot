# Upgrades and Migrator Design

<p className="subtitle">This doc is based on a preface for [RFC 850 WIP: Improving The Upgrade
Experience](https://docs.google.com/document/d/1ne8ai60iQnZfaYuB7QLDMIWgU5188Vn4_HBeUQ3GASY/edit)
and acts as a summary of the current design of our upgrade and database schema
management design. Subsequently additions have been made. The docs initial aim was to provide general information about the
relationship of our migrator service to our deployment types and
migrator's dependencies during the release process.</p>

## Migrator Overview

The `migrator` service is a short-lived container responsible for managing
Sourcegraph's databases (`pgsql` (*also referred to as frontend*), `codeintel-db`,
and `codeinsights-db`), and running schema migrations during startup and upgrades.

Its design accounts for  various unique characteristics of
versioning and database management at Sourcegraph. Specifically graphical
schema migrations, out-of-band migrations, and periodic schema migration squashing.

Sourcegraph utilizes a [directed acyclic graph](https://github.com/sourcegraph/sourcegraph/pull/30664)
of migration definitions, rather than a linear chain. In Sourcegraph's early days when schema migrations
were applied linearly, schema changes were frequent enough that schema changes generally conflicted with
the master branch by the time a PR passed CI. Moving to a graph of migrations means, devs won't need to
worry about other teammates concurrent schema changes unless they are working on the same table.

Similarly [squashing](#squashing-migrations) of schema migrations into a root definition reduced the number of migrations run on startup,
alleviating a common issue in which frequent transaction locks caused failed migration on Sourcegraph startup.
You can learn more in our [migrations overview docs](/migrations_overview#in-band-migrations).
Information on out of bound migrations can also be found there.


Migrator with its relevant artifacts in the sourcegraph/sourcegraph repo can be viewed as an orchestrator with two special functions --

1.  Migrator constructs migration plans given version ranges and a table
    of migrations which have been successfully applied (each schema has
    a table to track applied migrations within that schema). This logic
    is supported by a variety of files generated during releases and
    depends on the parent/child metadata of migrations generated via the
    sg migration tool.

2.  Migrator manages out-of-band migrations. These are data migrations
    that must be run within specific schema boundaries. Running OOB
    migrations at/after the deprecated version is unsupported. Migrator
    ensures that the necessary OOB migrations are run at [stopping
    points](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/internal/oobmigration/upgrade.go?L32-42)
    in a multiversion upgrade -- learn more
    [here](https://about.sourcegraph.com/blog/introducing-migrator-service).

### CLI design

Migrator is designed with a CLI tool interface -- taking various commands to alter
the state of the database and apply necessary schema changes. This
design was initially implemented as a tool for TS team members to assist
in multiversion upgrades and because it could easily be included over
multiple deployment methods as a containerized process run separately
from the frontend during startup ([which originally caused
issues](https://about.sourcegraph.com/blog/introducing-migrator-service)).
This replaced earlier go/migrate based strategies which ran in the
frontend on startup. While the migrator can operate as a CLI tool, it's
containerized as if it was another application which allows it to be run
as an initContainer (in Kubernetes deployments) or as a dependent
service (for Docker Compose deployments) to ensure that the necessary
migrations are applied before startup of the application proper.
Check out [RFC 469](https://docs.google.com/document/d/1_wqvNbEGhMzu45CHZNrLQYys7gl0AtVpLluGp4cbQKk/edit#heading=h.ddeuyk4t99yx) to learn more.

### Important Commands

The most important migrator commands are `up` and `upgrade`, with a notable mention to `drift`:

-   **Up** : The default command of migrator, `up` ensures that
    for a given version of the migrator, every migration defined at
    that build has been successfully applied in the connected database,
    this is specifically important to ensure *patch version migrations are run*.
    ***This must be run with the target version migrator image/build**.*
    If migrator and frontend pods are deployed in version lockstep, `up`
    ensures that ALL migrations required by the frontend will be successfully applied prior to boot.
    This is a syntactic sugar over a more internal `upto` command.
-   **Upgrade**: `upgrade` runs all migrations defined between two minor versions
    of Sourcegraph and requires that other services which may access the
    database are brought down. Before running, the database is checked
    for and schema drifts in order to prevent a failure while attempting
    a migration. `upgrade` relies on `stitched-migration-graph.json`.
-   **Drift**: This command pulls runs a diff between the current database
    schema and an expected definition packaged in migrator during the
    release process. Many migrator operations run this check before
    proceeding to ensure the database is in the expected state.

In general, the `up` command can be thought of as a **standard** upgrade
(rolling upgrades with no downtime) while the upgrade command is what
enables **multiversion** upgrades. In part, `up` was designed to maintain
our previous upgrade policy and is thus run as an initContainer (or
initContainer-like mechanism) of the frontend, i.e. between two versions
of Sourcegraph, the antecedent Sourcegraph services will continue to
work after the consequent version's migrations have been applied.

## Current Startup Dependency


![Migrator Startup Dependency](https://storage.googleapis.com/sourcegraph-assets/Docs/migrator-startup.png)

All of our deployment types currently utilize migrator during startup.
The frontend service won't start until migrator has been run with the
default up command. The frontend service will also validate the expected
schema (and OOB migration progress), and die on startup if this
validation pass fails. This ensures that the expected migrations for the
version in question have been run.

In docker-compose (*see diagram)*, this is accomplished via a chain of
`depends_on` clauses in the docker-compose.yaml
([link](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph-docker/-/blob/docker-compose/docker-compose.yaml?L217-223)).

**For our k8s based deployments (including the AMIs) migrator is run as
an initContainer within the frontend utilizing the up command on the
given pods startup.**

-   [Helm Ex](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph-helm/-/blob/charts/sourcegraph/templates/frontend/sourcegraph-frontend.Deployment.yaml?L49-76)
-   [Kustomize Ex](https://sourcegraph.com/github.com/sourcegraph/deploy-sourcegraph-k8s/-/blob/base/sourcegraph/frontend/sourcegraph-frontend.Deployment.yaml?L30-48)


## Auto-upgrade

Migrator has been incrementally improved over the last year in an
attempt to get closer and closer to auto-upgrades. After migrator v5.0.0
logic was added to the `pgsql` database and `frontend`/`frontend-internal` service to
attempt an automatic upgrade to the latest version of Sourcegraph on the startup of the frontend.

For more information about how this works see the
[docs](https://docs.sourcegraph.com/admin/updates/automatic#automatic-multi-version-upgrades).
Some notable points:

-   The upgrade operations in this case are
    [triggered](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@cf85b5803d32a91425f243930a4f50364625bcd2/-/blob/cmd/frontend/internal/cli/serve_cmd.go?L94-96)
    and
    [run](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@cf85b5803d32a91425f243930a4f50364625bcd2/-/blob/cmd/frontend/internal/cli/autoupgrade.go?L37-145)
    by the frontend container.

-   Migrator [looks
    for](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@cf85b5803d32a91425f243930a4f50364625bcd2/-/blob/internal/database/migration/cliutil/up.go?L125-128)
    the existence of the env var SRC_AUTOUPGRADE=true on services
    `sourcegraph-frontend`, `sourcegraph-frontend-internal`, and `migrator`.
    Otherwise it [looks in the frontend
    db](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@cf85b5803d32a91425f243930a4f50364625bcd2/-/blob/internal/database/migration/cliutil/up.go?L120-123)
    for the value of the autoupgrade column. These checks are performed
    with either the up or upgrade commands defined on the migrator.

-   The internal connections package to the DB now uses a special
    [sentinel
    value](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/internal/database/dbconn/connect.go?L31-37)
    to make connection attempts sleep if migrations are in progress.

-   A limited frontend is
    [served](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@cf85b5803d32a91425f243930a4f50364625bcd2/-/blob/cmd/frontend/internal/cli/autoupgrade.go?L78-88)
    by the frontend during an autoupgrade, displaying progress of the
    upgrade and any drift encountered.

-   All autoupgrades hit the multiversion upgrade endpoint and assume downtime for all Sourcegraph services besides the migrator and dbs.

## Migrator Release Artifacts

During the release of migrator we construct and build some artifacts
used by migrator to support its operations. Different artifacts must be
generated depending on the release type --

-   **Major**

    -   [lastMinorVersionInMajorRelease](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/internal/oobmigration/version.go?L84-87):
        Used to evaluate what oobmigrations must run, must be updated
        every major release. This essentially tells us when a minor
        version becomes a major version. *It may be useful elsewhere at
        some point.*

-   **Minor**

    -   [maxVersionString](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@5851edf/-/blob/internal/database/migration/shared/data/cmd/generator/consts.go?L12I):
        Defined in `consts.go` this string is used to tell migrator the
        latest **minor** version targetable for MVU and oobmigrations.
        [If not
        updated](https://github.com/sourcegraph/sourcegraph/issues/55048)
        multiversion upgrades cannot target the latest release. *Note
        this is used to determine how many versions should be included
        in the `stitched-migration-graph.json` file.*

    -   [Stitched-migration.json](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@v5.2.0/-/blob/internal/database/migration/shared/data/stitched-migration-graph.json):
        Used by multiversion upgrades to unsquash migrations. Generated
        during release
        [here](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/release/src/release.ts?L1101-1110).
        Learn more below.

-   **Patch**

    -   [Git_versions](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/cmd/migrator/generate.sh?L69-79):
        Defined in `generate.sh` this string array contains versions of
        Sourcegraph whose schemas should be embedded in migrator during a
        migrator build to enable drift detection without having to pull them
        directly from GitHub or, for older versions, from a pre-prepared GCS
        bucket (this is necessary in air gapped environments). This should
        be kept up to date with maxVersionString. [Learn
        more](https://github.com/sourcegraph/sourcegraph/issues/49813).

    -   [Squashed.sql](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@v5.2.0/-/blob/migrations/frontend/squashed.sql):
        for each database we
        [generate](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/internal/database/gen.sh?L18-20)
        a new squashed.sql file. It is
        [used](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@v5.2.0/-/blob/internal/database/migration/drift/util_search.go?L12-31)
        to help suggest fixes for certain types of database drift. For
        example for a missing database column [this
        search](https://sourcegraph.com/search?patternType=regexp&q=repo%3A%5Egithub%5C.com%2Fsourcegraph%2Fsourcegraph%24%40v5.0.6+file%3A%5Emigrations%2Ffrontend%2Fsquashed%5C.sql%24+%28%5E%7C%5Cb%29CREATE%5CsTABLE%5Csexternal_service_sync_jobs%28%24%7C%5Cb%29+OR+%28%5E%7C%5Cb%29ALTER%5CsTABLE%5CsONLY%5Csexternal_service_sync_jobs%28%24%7C%5Cb%29&groupBy=path)
        is used to suggest a definition.
    -   [Schema
        descriptions](https://raw.githubusercontent.com/sourcegraph/sourcegraph/v5.2.0/internal/database/schema.json):
        schema descriptions are
        [embedded](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@v5.2.0/-/blob/cmd/migrator/generate.sh?L64-66)
        in migrator on each new build as a reference for the expected schema
        state during drift checking.

### Stitched Migration JSON and Squashing Migrations

![migration graph](https://storage.googleapis.com/sourcegraph-assets/Docs/migration%20graph.png)

*\^\^ Generated with `sg migrations visualize --db frontend` on v5.2.0*

#### Squashing Migrations

[Periodically we
squash](https://github.com/sourcegraph/sourcegraph/pull/41819)
all migrations from two versions behind the current version into a
single migration using sg migration squash. This reduces the time
required to initialize a new database. This means that the migrator
image is built with a set of definitions embedded that doesn't reflect
the definition set in older versions. For multiversion upgrades this
presents a problem. To get around this, on minor releases we generate a
`stitched-migration-graph.json` file. Reference links:
[Bazel](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/release/src/util.ts?L293-327),
[Embed](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/internal/database/migration/shared/embed.go?L15-22),
[Generator](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/internal/database/migration/shared/data/cmd/generator/main.go)

####  Stitched Migration Graph

`stitched-migrations-graph.json` [stitches](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@v5.2.0/-/blob/internal/database/migration/stitch/stitch.go) (you can think of this as unsquashing) historic migrations using git magic, enabling the migrator to have a reference of older migrations. This serves a few purposes:

1.  When jumping across multiple versions we do not have access to a
    full record of migration definitions on migrator disk because some
    migrations will likely have been squashed. Therefore we need a way
    to ensure we don't miss migrations on a squash boundary. *Note we
    can't just re-apply the root migration after a squash because some
    schema state that\'s already represented in the root migration. This
    means the squashed root migration isn't always idempotent.*

2.  During a multiversion upgrade migrator must schedule out of band
    migrations to be started at some version and completed before
    upgrading past some later version. Migrator needs access to the
    unsquashed migration definitions to know which migrations must have
    run at the time the oob migration is triggered.

In standard/`up` upgrades `stitched-migrations.json` isn't necessary. This
is because `up` determines migrations to run by [comparing
migrations](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/internal/database/migration/definition/definition.go?L214-244)
listed as already run in the relevant db's `migration_logs` table directly
to those migration definitions
[embedded](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/migrations/embed.go)
in the migrator disk at build time for the current version, and running
any which haven't been run. We never squash away the previous minor
version of Sourcegraph, in this way we can guarantee the `migration_logs`
table migrations always has migrations in common with the migration
definitions on disk.

### Database Drifts

Database drift is any difference between the expected state of the
database at a given version and the actual state of the database. How
does it come to be? We've observed drift in customer databases for a
number of reasons, both due to admin tampering and problems in our own
release process:

-   **A backup/restore process that didn't include all objects**: This
    notably happened in a customer production instance causing their
    database to have no indexes, primary keys, or unique constraints defined.
-   **Modifying the database explicitly**: These are untracked manual
    changes to the database schema, observed occasionally in the wild,
    and in our dotcom deployment.
-   **Migration failures:** that occur during multi-version upgrade or
    the up command will cause the set of *successfully applied
    migrations* to be between two versions, where drift is well-defined.
-   **Site Admins Error:** Errors in git-ops like deploying to
    production on the wrong version of Sourcegraph manifests have
    introduced drift. Another source is the incorrect procedure in
    downgrading.
-   **Historic Bugs**: We, at one point, [too eagerly backfilled
    records that we should've instead
    applied](https://github.com/sourcegraph/sourcegraph/pull/55650).
    This bug was the result of changes being backported to the metadata
    definition of a migrations parent migrations, violating assumptions
    made during the generation of the stitched-migration-graph.json.

Database drift existing at the time of a migration can cause migrations
to fail when they try to reference some table property that is not in
the expected state. Not to mention the application may not behave as
expected if drift is present. Migrator includes a `drift` command intended
to help admins and CS team members to diagnose and resolve drift in
customer instances. Multiversion upgrades in particular check for drift
before starting unless the `--skip-drift-check argument` is supplied.

### Implementation Details

#### Versions & Runner

On startup the migrator service creates a
[runner](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@30d4d1dd457cde87c863a6d05cbcc0444025ed96/-/blob/cmd/migrator/shared/main.go?L31-36).

The `runner` is responsible for connecting to the given databases and
[running](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@4728edb797bc9affdbac940821c9e98c3fde2430/-/blob/internal/database/migration/runner/run.go)
any schema migrations defined in the embedded `migrations` directory
via the `up` entry command.

A `runner` [infers](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@4728edb797bc9affdbac940821c9e98c3fde2430/-/blob/internal/database/migration/schemas/schemas.go) the expected state of the database from schema definitions
[embeded in migrator](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@30d4d1dd457cde87c863a6d05cbcc0444025ed96/-/blob/migrations/embed.go)
when a migrator image is compiled. What this means is that migrator's concept of version,
 is the set of migrations defined in the `migrations` at compile time. In this way the `up` command easily facilitates dev versions of database schemas.

This "version" definition at compile time also tightly binds migrators concept of "version" to a given tag of Sourcegraph.
**The `up` command will only initialize a version of Sourcegraph, when the `migrator`
used to run `up` is the tagged version associated with the desired Sourcegraph version.**
For this reason a later version of `migrator` cannot be used to initialize an earlier
version of Sourcegraph.

For example, you use the latest migrator release `v5.6.9` to run the `upgrade` command bringing your databases from
`v4.2.0` to `v5.6.3`, rather than `v5.6.9`. Your security team hasn't approved images past this point. The upgrade command will have applied OOB migrations and schema migrations defined up to `v5.6.0`, the last minor release. To start your image you'll need to run migrator
`up` using the `v5.6.3` image, this will apply any schema migrations which may have been defined in the patch releases up to `v5.6.3` and thus existent in the embeded from `migrations` directory at the time of migrators compilation.

#### Migration Plan

While the `up` command's concept of version is a set of embedded definitions --
the `upgrade` command does have a concept of schema migrations associated to
version. This is the `stitched-migration-graph.json`. This file is generated on
minor releases of Sourcegraph, and defines migrations expected to have been run
at each minor version. This is necessary for two reasons --

1.  The root migration defined in the `migration` directory is a squashed
    migration, meaning, it represents many migrations composed into a single
    sql statement.
2.  Out of Bound migrations are triggered at a given version, and must complete
    before the schema is changed in some subsequent version.

This means that when applying migrations defined accross multiple versions, migrator
must stop and wait for OOB migrations to complete. To do this it needs to
know which migrations should have run at a given stopping point, which may have been
obscured by a subsequent squashing operation. This is where the `stitched-migration-graph.json`
file comes into play. It defines the set of migrations that should have been run at
a given minor version. Helping to construct a "[migration plan](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@4728edb797bc9affdbac940821c9e98c3fde2430/-/blob/internal/database/migration/multiversion/plan.go)" or path for `runner` to traverse.

The `stitched-migration.json` file is [generated](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@4728edb797bc9affdbac940821c9e98c3fde2430/-/blob/internal/database/migration/stitch/stitch.go) on every minor release, and is informed
by the state of the acyclic graph of migrations defined in the `migration` directory.
