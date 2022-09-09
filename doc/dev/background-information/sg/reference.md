<!-- DO NOT EDIT: generated via: go generate ./dev/sg -->

# sg reference

sg - The Sourcegraph developer tool!

Learn more: https://docs.sourcegraph.com/dev/background-information/sg

```sh
sg [GLOBAL FLAGS] command [COMMAND FLAGS] [ARGUMENTS...]
```

Global flags:

* `--config, -c="<value>"`: load sg configuration from `file` (default: sg.config.yaml)
* `--disable-analytics`: disable event logging (logged to '~/.sourcegraph/events')
* `--disable-output-detection`: use fixed output configuration instead of detecting terminal capabilities
* `--disable-overwrite`: disable loading additional sg configuration from overwrite file (see -overwrite)
* `--overwrite, -o="<value>"`: load sg configuration from `file` that is gitignored and can be used to, for example, add credentials (default: sg.config.overwrite.yaml)
* `--skip-auto-update`: prevent sg from automatically updating itself
* `--verbose, -v`: toggle verbose mode


## sg start

🌟 Starts the given commandset. Without a commandset it starts the default Sourcegraph dev environment.

Use this to start your Sourcegraph environment!

Available comamndsets in `sg.config.yaml`:

* api-only
* batches 🦡
* codeintel
* dotcom
* enterprise
* enterprise-codeinsights
* enterprise-codeintel 🧠
* enterprise-e2e
* iam
* monitoring
* monitoring-alerts
* oss
* oss-web-standalone
* oss-web-standalone-prod
* otel
* web-standalone
* web-standalone-prod

```sh
# Run default environment, Sourcegraph enterprise:
$ sg start

# List available environments (defined under 'commandSets' in 'sg.config.yaml'):
$ sg start -help

# Run the enterprise environment with code-intel enabled:
$ sg start enterprise-codeintel

# Run the environment for Batch Changes development:
$ sg start batches

# Override the logger levels for specific services
$ sg start --debug=gitserver --error=enterprise-worker,enterprise-frontend enterprise
```

Flags:

* `--crit, -c="<value>"`: Services to set at info crit level.
* `--debug, -d="<value>"`: Services to set at debug log level.
* `--error, -e="<value>"`: Services to set at info error level.
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--info, -i="<value>"`: Services to set at info log level.
* `--warn, -w="<value>"`: Services to set at warn log level.

## sg run

Run the given commands.

Runs the given command. If given a whitespace-separated list of commands it runs the set of commands.

Available commands in `sg.config.yaml`:

* batches-executor
* batches-executor-firecracker
* bext
* caddy
* codeintel-executor
* codeintel-worker
* debug-env: Debug env vars
* docsite: Docsite instance serving the docs
* executor-template
* frontend: Enterprise frontend
* github-proxy
* gitserver
* grafana
* jaeger
* loki
* minio
* monitoring-generator
* oss-frontend
* oss-repo-updater
* oss-symbols
* oss-web: Open source version of the web app
* oss-worker
* otel-collector: OpenTelemetry collector
* postgres_exporter
* prometheus
* redis-postgres: Dockerized version of redis and postgres
* repo-updater
* searcher
* server: Run an all-in-one sourcegraph/server image
* storybook
* symbols
* syntax-highlighter
* web-standalone-http-prod: Standalone web frontend (production) with API proxy to a configurable URL
* web-standalone-http: Standalone web frontend (dev) with API proxy to a configurable URL
* web: Enterprise version of the web app
* worker
* zoekt-index-0
* zoekt-index-1
* zoekt-indexserver-template
* zoekt-web-0
* zoekt-web-1
* zoekt-web-template

```sh
# Run specific commands:
$ sg run gitserver
$ sg run frontend

# List available commands (defined under 'commands:' in 'sg.config.yaml'):
$ sg run -help

# Run multiple commands:
$ sg run gitserver frontend repo-updater
```

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

## sg ci

Interact with Sourcegraph's Buildkite continuous integration pipelines.

Note that Sourcegraph's CI pipelines are under our enterprise license: https://github.com/sourcegraph/sourcegraph/blob/main/LICENSE.enterprise

```sh
# Preview what a CI run for your current changes will look like
$ sg ci preview

# Check on the status of your changes on the current branch in the Buildkite pipeline
$ sg ci status
# Check on the status of a specific branch instead
$ sg ci status --branch my-branch
# Block until the build has completed (it will send a system notification)
$ sg ci status --wait
# Get status for a specific build number
$ sg ci status --build 123456

# Pull logs of failed jobs to stdout
$ sg ci logs
# Push logs of most recent main failure to local Loki for analysis
# You can spin up a Loki instance with 'sg run loki grafana'
$ sg ci logs --branch main --out http://127.0.0.1:3100
# Get the logs for a specific build number, useful when debugging
$ sg ci logs --build 123456

# Manually trigger a build on the CI with the current branch
$ sg ci build
# Manually trigger a build on the CI on the current branch, but with a specific commit
$ sg ci build --commit my-commit
# Manually trigger a main-dry-run build of the HEAD commit on the current branch
$ sg ci build main-dry-run
$ sg ci build --force main-dry-run
# Manually trigger a main-dry-run build of a specified commit on the current ranch
$ sg ci build --force --commit my-commit main-dry-run
# View the available special build types
$ sg ci build --help
```

### sg ci preview

Preview the pipeline that would be run against the currently checked out branch.


Flags:

* `--branch, -b="<value>"`: Branch `name` of build to target (defaults to current branch)
* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg ci status

Get the status of the CI run associated with the currently checked out branch.


Flags:

* `--branch, -b="<value>"`: Branch `name` of build to target (defaults to current branch)
* `--build, -n="<value>"`: Override branch detection with a specific build `number`
* `--commit, -c="<value>"`: Override branch detection with the latest build for `commit`
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--pipeline, -p="<value>"`: Select a custom Buildkite `pipeline` in the Sourcegraph org (default: sourcegraph)
* `--view, -v`: Open build page in browser
* `--wait, -w`: Wait by blocking until the build is finished

### sg ci build

Manually request a build for the currently checked out commit and branch (e.g. to trigger builds on forks or with special run types).

Optionally provide a run type to build with.

This command is useful when:

- you want to trigger a build with a particular run type, such as 'main-dry-run'
- triggering builds for PRs from forks (such as those from external contributors), which do not trigger Buildkite builds automatically for security reasons (we do not want to run insecure code on our infrastructure by default!)

Supported run types when providing an argument for 'sg ci build [runtype]':

* main-dry-run
* docker-images-patch
* docker-images-patch-notest
* docker-images-candidates-notest
* executor-patch-notest
* backend-integration

For run types that require branch arguments, you will be prompted for an argument, or you
can provide it directly (for example, 'sg ci build [runtype] [argument]').

Learn more about pipeline run types in https://docs.sourcegraph.com/dev/background-information/ci/reference.

Arguments: `[runtype]`

Flags:

* `--commit, -c="<value>"`: `commit` from the current branch to build (defaults to current commit)
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--pipeline, -p="<value>"`: Select a custom Buildkite `pipeline` in the Sourcegraph org (default: sourcegraph)

### sg ci logs

Get logs from CI builds (e.g. to grep locally).

Get logs from CI builds, and output them in stdout or push them to Loki. By default only gets failed jobs - to change this, use the '--state' flag.

The '--job' flag can be used to narrow down the logs returned - you can provide either the ID, or part of the name of the job you want to see logs for.

To send logs to a Loki instance, you can provide --out=http://127.0.0.1:3100 after spinning up an instance with 'sg run loki grafana'.
From there, you can start exploring logs with the Grafana explore panel.



Flags:

* `--branch, -b="<value>"`: Branch `name` of build to target (defaults to current branch)
* `--build, -n="<value>"`: Override branch detection with a specific build `number`
* `--commit, -c="<value>"`: Override branch detection with the latest build for `commit`
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--job, -j="<value>"`: ID or name of the job to export logs for
* `--out, -o="<value>"`: Output `format`: one of [terminal|simple|json], or a URL pointing to a Loki instance, such as http://127.0.0.1:3100 (default: terminal)
* `--overwrite-state="<value>"`: `state` to overwrite the job state metadata
* `--pipeline, -p="<value>"`: Select a custom Buildkite `pipeline` in the Sourcegraph org (default: sourcegraph)
* `--state, -s="<value>"`: Job `state` to export logs for (provide an empty value for all states) (default: failed)

### sg ci docs

Render reference documentation for build pipeline types.

An online version of the rendered documentation is also available in https://docs.sourcegraph.com/dev/background-information/ci/reference.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg ci open

Open Sourcegraph's Buildkite page in browser.

Arguments: `[pipeline]`

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg ci search-failures

Open Sourcegraph's CI failures Grafana logs page in browser.

Arguments: `[text to search for]`

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--step="<value>"`: Filter by step name (--step STEP_NAME will translate to '.*STEP_NAME.*')

## sg test

Run the given test suite.

Testsuites are defined in sg configuration.

Available testsuites in `sg.config.yaml`:

* backend
* backend-integration
* bext
* bext-build
* bext-e2e
* bext-integration
* client
* client-e2e
* client-regression
* docsite
* web-integration

```sh
# Run different test suites:
$ sg test backend
$ sg test backend-integration
$ sg test client
$ sg test client-e2e

# List available test suites:
$ sg test -help

# Arguments are passed along to the command
$ sg test backend-integration -run TestSearch
```

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

## sg lint

Run all or specified linters on the codebase.

To run all checks, don't provide an argument. You can also provide multiple arguments to run linters for multiple targets.

```sh
# Run all possible checks
$ sg lint

# Run only go related checks
$ sg lint go

# Run only shell related checks
$ sg lint shell

# Run only client related checks
$ sg lint client

# List all available check groups
$ sg lint --help
```

Flags:

* `--annotations`: Write helpful output to ./annotations directory
* `--fail-fast, --ff`: Exit immediately if an issue is encountered (not available with '-fix')
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--fix, -f`: Try to fix any lint issues

### sg lint urls

Check for broken urls in the codebase.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg lint go

Check go code for linting errors, forbidden imports, generated files, etc.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg lint docs

Documentation checks.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg lint dockerfiles

Check Dockerfiles for Sourcegraph best practices.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg lint client

Check client code for linting errors, forbidden imports, etc.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg lint svg

Check svg assets.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg lint shell

Check shell code for linting errors, formatting, etc.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

## sg generate

Run code and docs generation tasks.

If no target is provided, all target are run with default arguments.

```sh
$ sg --verbose generate ... # Enable verbose output
```

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--quiet, -q`: Suppress all output but errors from generate tasks

### sg generate go

Run go generate [packages...] on the codebase.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

## sg db

Interact with local Sourcegraph databases for development.

```sh
# Reset the Sourcegraph 'frontend' database
$ sg db reset-pg

# Reset the 'frontend' and 'codeintel' databases
$ sg db reset-pg -db=frontend,codeintel

# Reset all databases ('frontend', 'codeintel', 'codeinsights')
$ sg db reset-pg -db=all

# Reset the redis database
$ sg db reset-redis

# Create a site-admin user whose email and password are foo@sourcegraph.com and sourcegraph.
$ sg db add-user -name=foo
```

### sg db reset-pg

Drops, recreates and migrates the specified Sourcegraph database.

If -db is not set, then the "frontend" database is used (what's set as PGDATABASE in env or the sg.config.yaml). If -db is set to "all" then all databases are reset and recreated.


Flags:

* `--db="<value>"`: The target database instance. (default: frontend)
* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg db reset-redis

Drops, recreates and migrates the specified Sourcegraph Redis database.

```sh
$ sg db reset-redis
```

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg db add-user

Create an admin sourcegraph user.

Run 'sg db add-user -username bob' to create an admin user whose email is bob@sourcegraph.com. The password will be printed if the operation succeeds


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--password="<value>"`: Password for user (default: sourcegraphsourcegraph)
* `--username="<value>"`: Username for user (default: sourcegraph)

## sg migration

Modifies and runs database migrations.

```sh
# Migrate local default database up all the way
$ sg migration up

# Migrate specific database down one migration
$ sg migration down --db codeintel

# Add new migration for specific database
$ sg migration add --db codeintel 'add missing index'

# Squash migrations for default database
$ sg migration squash
```

### sg migration add

Add a new migration file.

Available schemas:

* frontend
* codeintel
* codeinsights

Arguments: `<name>`

Flags:

* `--db="<value>"`: The target database `schema` to modify (default: frontend)
* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg migration revert

Revert the migrations defined on the given commit.

Available schemas:

* frontend
* codeintel
* codeinsights

Arguments: `<commit>`

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg migration up

Apply all migrations.

Available schemas:

* frontend
* codeintel
* codeinsights

```sh
$ sg migration up [-db=<schema>]
```

Flags:

* `--db="<value>"`: The target `schema(s)` to modify. Comma-separated values are accepted. Supply "all" to migrate all schemas. (default: [all])
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--ignore-single-dirty-log`: Ignore a single previously failed attempt if it will be immediately retried by this operation.
* `--ignore-single-pending-log`: Ignore a single pending migration attempt if it will be immediately retried by this operation.
* `--noop-privileged`: Skip application of privileged migrations, but record that they have been applied. This assumes the user has already applied the required privileged migrations with elevated permissions.
* `--privileged-hash="<value>"`: Running -noop-privileged without this value will supply a value that will unlock migration application for the current upgrade operation. Future (distinct) upgrade operations will require a unique hash.
* `--skip-oobmigration-validation`: Do not attempt to validate the progress of out-of-band migrations.
* `--skip-upgrade-validation`: Do not attempt to compare the previous instance version with the target instance version for upgrade compatibility. Please refer to https://docs.sourcegraph.com/admin/updates#update-policy for our instance upgrade compatibility policy.
* `--unprivileged-only`: Refuse to apply privileged migrations.

### sg migration upto

Ensure a given migration has been applied - may apply dependency migrations.

Available schemas:

* frontend
* codeintel
* codeinsights

```sh
$ sg migration upto -db=<schema> -target=<target>,<target>,...
```

Flags:

* `--db="<value>"`: The target `schema` to modify.
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--ignore-single-dirty-log`: Ignore a single previously failed attempt if it will be immediately retried by this operation.
* `--ignore-single-pending-log`: Ignore a single pending migration attempt if it will be immediately retried by this operation.
* `--noop-privileged`: Skip application of privileged migrations, but record that they have been applied. This assumes the user has already applied the required privileged migrations with elevated permissions.
* `--privileged-hash="<value>"`: Running -noop-privileged without this value will supply a value that will unlock migration application for the current upgrade operation. Future (distinct) upgrade operations will require a unique hash.
* `--target="<value>"`: The `migration` to apply. Comma-separated values are accepted.
* `--unprivileged-only`: Refuse to apply privileged migrations.

### sg migration undo

Revert the last migration applied - useful in local development.

Available schemas:

* frontend
* codeintel
* codeinsights

```sh
$ sg migration undo -db=<schema>
```

Flags:

* `--db="<value>"`: The target `schema` to modify.
* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg migration downto

Revert any applied migrations that are children of the given targets - this effectively "resets" the schema to the target version.

Available schemas:

* frontend
* codeintel
* codeinsights

```sh
$ sg migration downto -db=<schema> -target=<target>,<target>,...
```

Flags:

* `--db="<value>"`: The target `schema` to modify.
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--ignore-single-dirty-log`: Ignore a single previously failed attempt if it will be immediately retried by this operation.
* `--ignore-single-pending-log`: Ignore a single pending migration attempt if it will be immediately retried by this operation.
* `--noop-privileged`: Skip application of privileged migrations, but record that they have been applied. This assumes the user has already applied the required privileged migrations with elevated permissions.
* `--privileged-hash="<value>"`: Running -noop-privileged without this value will supply a value that will unlock migration application for the current upgrade operation. Future (distinct) upgrade operations will require a unique hash.
* `--target="<value>"`: The migration to apply. Comma-separated values are accepted.
* `--unprivileged-only`: Refuse to apply privileged migrations.

### sg migration validate

Validate the current schema.

Available schemas:

* frontend
* codeintel
* codeinsights


Flags:

* `--db="<value>"`: The target `schema(s)` to validate. Comma-separated values are accepted. Supply "all" to validate all schemas. (default: [all])
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--skip-out-of-band-migrations`: Do not attempt to validate out-of-band migration status.

### sg migration describe

Describe the current database schema.

Available schemas:

* frontend
* codeintel
* codeinsights


Flags:

* `--db="<value>"`: The target `schema` to describe.
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--force`: Force write the file if it already exists.
* `--format="<value>"`: The target output format.
* `--no-color`: If writing to stdout, disable output colorization.
* `--out="<value>"`: The file to write to. If not supplied, stdout is used.

### sg migration drift

Detect differences between the current database schema and the expected schema.

Available schemas:

* frontend
* codeintel
* codeinsights


Flags:

* `--db="<value>"`: The target `schema` to compare.
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--file="<value>"`: The target schema description file.
* `--version="<value>"`: The target schema version. Must be resolvable as a git revlike on the Sourcegraph repository.

### sg migration add-log

Add an entry to the migration log.

Available schemas:

* frontend
* codeintel
* codeinsights

```sh
$ sg migration add-log -db=<schema> -version=<version> [-up=true|false]
```

Flags:

* `--db="<value>"`: The target `schema` to modify.
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--up`: The migration direction.
* `--version="<value>"`: The migration `version` to log. (default: 0)

### sg migration leaves

Identify the migration leaves for the given commit.

Available schemas:

* frontend
* codeintel
* codeinsights

Arguments: `<commit>`

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg migration squash

Collapse migration files from historic releases together.

Available schemas:

* frontend
* codeintel
* codeinsights

Arguments: `<current-release>`

Flags:

* `--db="<value>"`: The target database `schema` to modify (default: frontend)
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--in-container`: Launch Postgres in a Docker container for squashing; do not use the host
* `--in-timescaledb-container`: Launch TimescaleDB in a Docker container for squashing; do not use the host
* `--skip-data`: Skip writing data rows into the squashed migration
* `--skip-teardown`: Skip tearing down the database created to run all registered migrations

### sg migration squash-all

Collapse schema definitions into a single SQL file.

Available schemas:

* frontend
* codeintel
* codeinsights


Flags:

* `--db="<value>"`: The target database `schema` to modify (default: frontend)
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--in-container`: Launch Postgres in a Docker container for squashing; do not use the host
* `--in-timescaledb-container`: Launch TimescaleDB in a Docker container for squashing; do not use the host
* `--skip-data`: Skip writing data rows into the squashed migration
* `--skip-teardown`: Skip tearing down the database created to run all registered migrations
* `-f="<value>"`: The output filepath

### sg migration visualize

Output a DOT visualization of the migration graph.

Available schemas:

* frontend
* codeintel
* codeinsights


Flags:

* `--db="<value>"`: The target database `schema` to modify (default: frontend)
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `-f="<value>"`: The output filepath

### sg migration rewrite

Rewrite schemas definitions as they were at a particular version.

Available schemas:

* frontend
* codeintel
* codeinsights


Flags:

* `--db="<value>"`: The target database `schema` to modify (default: frontend)
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--rev="<value>"`: The target revision

## sg insights

Tools to interact with Code Insights data.


### sg insights decode-id

Decodes an encoded insight ID found on the frontend into a view unique_id.

Run 'sg insights decode-id' to decode 1+ frontend IDs which can then be used for SQL queries


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg insights series-ids

Gets all insight series ID from the base64 encoded frontend ID.

Run 'sg insights series-ids' to decode a frontend ID and find all related series IDs


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

## sg telemetry

Operations relating to Sourcegraph telemetry.


### sg telemetry allowlist

Edit the usage data allow list.


Utility that will generate SQL to add and remove events from the usage data allow list.
https://docs.sourcegraph.com/dev/background-information/data-usage-pipeline#allow-list

Events are keyed by event name and passed in as additional arguments to the add and remove subcommands.


```sh
# Generate SQL to add events from the allow list
$ sg telemetry allowlist add EVENT_ONE EVENT_TWO

# Generate SQL to remove events from the allow list
$ sg telemetry allowlist remove EVENT_ONE EVENT_TWO

# Automatically generate migration files associated with the allow list modification
$ sg telemetry allowlist add --migration EVENT_ONE EVENT_TWO

# Provide a specific migration name for the migration files
$ sg telemetry allowlist add --migration --name my_migration_name EVENT_ONE EVENT_TWO
```

#### sg telemetry allowlist add

Generate the SQL required to add events to the allow list.

```sh
# Generate SQL to add events from the allow list
$ sg telemetry allowlist add EVENT_ONE EVENT_TWO

# Automatically generate migration files associated with the allow list modification
$ sg telemetry allowlist add --migration EVENT_ONE EVENT_TWO

# Provide a specific migration name for the migration files
$ sg telemetry allowlist add --migration --name my_migration_name EVENT_ONE EVENT_TWO
```

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--migration`: Create migration files with the generated SQL.
* `--name="<value>"`: Specifies the name of the resulting migration. (default: sg_telemetry_allowlist)

#### sg telemetry allowlist remove

Generate the SQL required to remove events from the allow list.

```sh
# Generate SQL to add events from the allow list
$ sg telemetry allowlist remove EVENT_ONE EVENT_TWO

# Automatically generate migration files associated with the allow list modification
$ sg telemetry allowlist remove --migration EVENT_ONE EVENT_TWO

# Provide a specific migration name for the migration files
$ sg telemetry allowlist remove --migration --name my_migration_name EVENT_ONE EVENT_TWO
```

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--migration`: Create migration files with the generated SQL.
* `--name="<value>"`: Specifies the name of the resulting migration. (default: sg_telemetry_allowlist)

## sg secret

Manipulate secrets stored in memory and in file.

```sh
# List all secrets stored in your local configuration.
$ sg secret list

# Remove the secrets associated with buildkite (sg ci build) - supports autocompletion for
# ease of use
$ sg secret reset buildkite
```

### sg secret reset

Remove a specific secret from secrets file.

Arguments: `<...key>`

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg secret list

List all stored secrets.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--view, -v`: Display configured secrets when listing

## sg setup

Validate and set up your local dev environment!.


Flags:

* `--check, -c`: Run checks and report setup state
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--fix, -f`: Fix all checks
* `--oss`: Omit Sourcegraph-teammate-specific setup

## sg teammate

Get information about Sourcegraph teammates.

For example, you can check a teammate's current time and find their handbook bio!

```sh
# Get the current time of a team mate based on their slack handle (case insensitive).
$ sg teammate time @dax
$ sg teammate time dax
# or their full name (case insensitive)
$ sg teammate time thorsten ball

# Open their handbook bio
$ sg teammate handbook asdine
```

### sg teammate time

Get the current time of a Sourcegraph teammate.

Arguments: `<nickname>`

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg teammate handbook

Open the handbook page of a Sourcegraph teammate.

Arguments: `<nickname>`

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

## sg rfc

List, search, and open Sourcegraph RFCs.

```sh
# List all RFCs
$ sg rfc list

# Search for an RFC
$ sg rfc search "search terms"

# Open a specific RFC
$ sg rfc open 420
```

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

## sg adr

List, search, view, and create Sourcegraph Architecture Decision Records (ADRs).

We use Architecture Decision Records (ADRs) only for logging decisions that have notable
architectural impact on our codebase. Since we're a high-agency company, we encourage any
contributor to commit an ADR if they've made an architecturally significant decision.

ADRs are not meant to replace our current RFC process but to complement it by capturing
decisions made in RFCs. However, ADRs do not need to come out of RFCs only. GitHub issues
or pull requests, PoCs, team-wide discussions, and similar processes may result in an ADR
as well.

Learn more about ADRs here: https://docs.sourcegraph.com/dev/adr

```sh
# List all ADRs
$ sg adr list

# Search for an ADR
$ sg adr search "search terms"

# Open a specific index
$ sg adr view 420

# Create a new ADR!
$ sg adr create my ADR title
```

### sg adr list

List all ADRs.


Flags:

* `--asc`: List oldest ADRs first
* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg adr search

Search ADR titles and content.

Arguments: `[terms...]`

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg adr view

View an ADR.

Arguments: `[number]`

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg adr create

Create an ADR!.

Arguments: `<title>`

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

## sg live

Reports which version of Sourcegraph is currently live in the given environment.

Prints the Sourcegraph version deployed to the given environment.

Available preset environments:

* s2
* dotcom
* k8s

```sh
# See which version is deployed on a preset environment
$ sg live s2
$ sg live dotcom
$ sg live k8s

# See which version is deployed on a custom environment
$ sg live https://demo.sourcegraph.com

# List environments
$ sg live -help

# Check for commits further back in history
$ sg live -n 50 s2
```

Flags:

* `--commits, -c, -n="<value>"`: Number of commits to check for live version (default: 20)
* `--feedback`: provide feedback about this command by opening up a GitHub discussion

## sg ops

Commands used by operations teams to perform common tasks.

Supports internal deploy-sourcegraph repos (non-customer facing)


### sg ops update-images

Updates images in given directory to latest published image.

Updates images in given directory to latest published image.
Ex: in deploy-sourcegraph-cloud, run `sg ops update-images base/.`

Arguments: `<dir>`

Flags:

* `--cr-password="<value>"`: `password` or access token for the container registry
* `--cr-username="<value>"`: `username` for the container registry
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--kind, -k="<value>"`: the `kind` of deployment (one of 'k8s', 'helm', 'compose') (default: k8s)
* `--pin-tag, -t="<value>"`: pin all images to a specific sourcegraph `tag` (e.g. '3.36.2', 'insiders') (default: latest main branch tag)

### sg ops inspect-tag

Inspect main branch tag details from a image or tag.

```sh
# Inspect a full image
$ sg ops inspect-tag index.docker.io/sourcegraph/cadvisor:159625_2022-07-11_225c8ae162cc@sha256:foobar

# Inspect just the tag
$ sg ops inspect-tag 159625_2022-07-11_225c8ae162cc

# Get the build number
$ sg ops inspect-tag -p build 159625_2022-07-11_225c8ae162cc
```

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--property, -p="<value>"`: only output a specific `property` (one of: 'build', 'date', 'commit')

## sg page

Page engineers at Sourcegraph - mostly used within scripts to automate paging alerts.

```sh
$ sg page --opsgenie.token [TOKEN] --message "something is broken" [my-schedule-on-ops-genie]
```

Flags:

* `--description, -d="<value>"`: Description for the paging alert (optional)
* `--escalation="<value>"`: Escalation team(s) to alert (if provided, target schedules can be omitted)
* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--message, -m="<value>"`: Message for the paging alert
* `--opsgenie.token="<value>"`: OpsGenie token
* `--priority, -p="<value>"`: Alert priority, importance decreases from P1 (critical) to P5 (lowest), defaults to P5 (default: P5)
* `--url="<value>"`: URL field for alert details (optional)

## sg help

Get help and docs about sg.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--full, -f`: generate full markdown sg reference
* `--help, -h`: show help
* `--output="<value>"`: write reference to `file`

## sg feedback

Provide feedback about sg.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

## sg version

View details for this installation of sg.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg version changelog

See what's changed in or since this version of sg.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--limit="<value>"`: Number of changelog entries to show. (default: 5)
* `--next`: Show changelog for changes you would get if you upgrade.

## sg update

Update local sg installation.

Update local sg installation with the latest changes. To see what's new, run:

    sg version changelog -next


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

## sg logo

Print the sg logo.

By default, prints the sg logo in different colors. When the 'classic' argument is passed it prints the classic logo.

Arguments: `[classic]`

Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

## sg analytics

Manage analytics collected by sg.


### sg analytics submit

Make sg better by submitting all analytics stored locally!.

Requires HONEYCOMB_ENV_TOKEN or OTEL_EXPORTER_OTLP_ENDPOINT to be set.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg analytics reset

Delete all analytics stored locally.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion

### sg analytics view

View all analytics stored locally.


Flags:

* `--feedback`: provide feedback about this command by opening up a GitHub discussion
* `--raw`: view raw data
