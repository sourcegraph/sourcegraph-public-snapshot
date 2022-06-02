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
* core-app
* dotcom
* enterprise
* enterprise-codeinsights
* enterprise-codeintel 🧠
* enterprise-e2e
* monitoring
* monitoring-alerts
* oss
* oss-web-standalone
* oss-web-standalone-prod
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
* debug-env
* docsite
* enterprise-frontend
* enterprise-repo-updater
* enterprise-symbols
* enterprise-web
* enterprise-worker
* executor-template
* frontend
* github-proxy
* gitserver
* grafana
* jaeger
* loki
* minio
* monitoring-generator
* postgres_exporter
* precise-code-intel-worker
* prometheus
* redis-postgres
* repo-updater
* searcher
* storybook
* symbols
* syntax-highlighter
* web
* web-standalone-http
* web-standalone-http-prod
* worker
* zoekt-indexserver-0
* zoekt-indexserver-1
* zoekt-indexserver-template
* zoekt-webserver-0
* zoekt-webserver-1
* zoekt-webserver-template

```sh
# Run specific commands:
$ sg run gitserver
$ sg run frontend

# List available commands (defined under 'commands:' in 'sg.config.yaml'):
$ sg run -help

# Run multiple commands:
$ sg run gitserver frontend repo-updater
```

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


### sg ci status

Get the status of the CI run associated with the currently checked out branch.


Flags:

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

### sg ci logs

Get logs from CI builds (e.g. to grep locally).

Get logs from CI builds, and output them in stdout or push them to Loki. By default only gets failed jobs - to change this, use the '--state' flag.

The '--job' flag can be used to narrow down the logs returned - you can provide either the ID, or part of the name of the job you want to see logs for.

To send logs to a Loki instance, you can provide --out=http://127.0.0.1:3100 after spinning up an instance with 'sg run loki grafana'.
From there, you can start exploring logs with the Grafana explore panel.



Flags:

* `--build="<value>"`: Override branch detection with a specific build `number`
* `--job, -j="<value>"`: ID or name of the job to export logs for
* `--out, -o="<value>"`: Output `format`: one of [terminal|simple|json], or a URL pointing to a Loki instance, such as http://127.0.0.1:3100 (default: terminal)
* `--overwrite-state="<value>"`: `state` to overwrite the job state metadata
* `--state, -s="<value>"`: Job `state` to export logs for (provide an empty value for all states) (default: failed)

### sg ci docs

Render reference documentation for build pipeline types.

An online version of the rendered documentation is also available in https://docs.sourcegraph.com/dev/background-information/ci/reference.


### sg ci open

Open Sourcegraph's Buildkite page in browser.

Arguments: `[pipeline]`

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
* docsite
* frontend
* frontend-e2e
* web-integration

```sh
# Run different test suites:
$ sg test backend
$ sg test backend-integration
$ sg test frontend
$ sg test frontend-e2e

# List available test suites:
$ sg test -help

# Arguments are passed along to the command
$ sg test backend-integration -run TestSearch
```

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

* `--annotations`: Write helpful output to annotations directory

### sg lint urls

Check for broken urls in the codebase.


### sg lint go

Check go code for linting errors, forbidden imports, generated files, etc.


### sg lint docs

Documentation checks.


### sg lint dockerfiles

Check Dockerfiles for Sourcegraph best practices.


### sg lint client

Check client code for linting errors, forbidden imports, etc.


### sg lint svg

Check svg assets.


### sg lint shell

Check shell code for linting errors, formatting, etc.


## sg generate

Run code and docs generation tasks.

If no target is provided, all target are run with default arguments.

```sh
$ sg --verbose generate ... # Enable verbose output
```

Flags:

* `--quiet, -q`: Suppress all output but errors from generate tasks

### sg generate go

Run go generate [packages...] on the codebase.


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

### sg db reset-redis

Drops, recreates and migrates the specified Sourcegraph Redis database.

```sh
$ sg db reset-redis
```

### sg db add-user

Create an admin sourcegraph user.

Run 'sg db add-user -name bob' to create an admin user whose email is bob@sourcegraph.com. The password will be printed if the operation succeeds


Flags:

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

### sg migration revert

Revert the migrations defined on the given commit.

Available schemas:

* frontend
* codeintel
* codeinsights

Arguments: `<commit>`

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
* `--ignore-single-dirty-log`: Ignore a previously failed attempt if it will be immediately retried by this operation.
* `--unprivileged-only`: Do not apply privileged migrations.

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
* `--ignore-single-dirty-log`: Ignore a previously failed attempt if it will be immediately retried by this operation.
* `--target="<value>"`: The `migration` to apply. Comma-separated values are accepted.
* `--unprivileged-only`: Do not apply privileged migrations.

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
* `--ignore-single-dirty-log`: Ignore a previously failed attempt if it will be immediately retried by this operation.

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
* `--ignore-single-dirty-log`: Ignore a previously failed attempt if it will be immediately retried by this operation.
* `--target="<value>"`: The migration to apply. Comma-separated values are accepted.
* `--unprivileged-only`: Do not apply privileged migrations.

### sg migration validate

Validate the current schema.

Available schemas:

* frontend
* codeintel
* codeinsights


Flags:

* `--db="<value>"`: The target `schema(s)` to modify. Comma-separated values are accepted. Supply "all" to migrate all schemas. (default: [all])

### sg migration describe

Describe the current database schema.

Available schemas:

* frontend
* codeintel
* codeinsights


Flags:

* `--db="<value>"`: The target `schema` to describe.
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
* `--version="<value>"`: The target schema version. Must be resolvable as a git revlike on the sourcegraph repository.

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
* `--up`: The migration direction.
* `--version="<value>"`: The migration `version` to log. (default: 0)

### sg migration leaves

Identiy the migration leaves for the given commit.

Available schemas:

* frontend
* codeintel
* codeinsights

Arguments: `<commit>`

### sg migration squash

Collapse migration files from historic releases together.

Available schemas:

* frontend
* codeintel
* codeinsights

Arguments: `<current-release>`

Flags:

* `--db="<value>"`: The target database `schema` to modify (default: frontend)
* `--in-container`: Launch Postgres in a Docker container for squashing; do not use the host
* `--skip-teardown`: Skip tearing down the database created to run all registered migrations

### sg migration squash-all

Collapse schema definitions into a single SQL file.

Available schemas:

* frontend
* codeintel
* codeinsights


Flags:

* `--db="<value>"`: The target database `schema` to modify (default: frontend)
* `--in-container`: Launch Postgres in a Docker container for squashing; do not use the host
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
* `-f="<value>"`: The output filepath

## sg doctor

Run checks to test whether system is in correct state to run Sourcegraph.

Arguments: `[...checks]`

## sg secret

Manipulate secrets stored in memory and in file.

```sh
# List all secrets stored in your local configuration.
$ sg secret list

# Remove the secrets associated with buildkite (sg ci build)
$ sg secret reset buildkite
```

### sg secret reset

Remove a specific secret from secrets file.

Arguments: `<...key>`

### sg secret list

List all stored secrets.


Flags:

* `--view, -v`: Display configured secrets when listing

## sg setup

Set up your local dev environment!.


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

### sg teammate handbook

Open the handbook page of a Sourcegraph teammate.

Arguments: `<nickname>`

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

## sg live

Reports which version of Sourcegraph is currently live in the given environment.

Prints the Sourcegraph version deployed to the given environment.

Available preset environments:

* cloud
* k8s

```sh
# See which version is deployed on a preset environment
$ sg live cloud
$ sg live k8s

# See which version is deployed on a custom environment
$ sg live https://demo.sourcegraph.com

# List environments:
$ sg live -help
```

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
* `--kind="<value>"`: the `kind` of deployment (one of 'k8s', 'helm') (default: k8s)
* `--pin-tag="<value>"`: pin all images to a specific sourcegraph `tag` (e.g. 3.36.2, insiders)

## sg analytics

Manage analytics collected by sg.


### sg analytics submit

Make sg better by submitting all analytics stored locally!.

Uses OKAYHQ_TOKEN, or fetches a token from gcloud or 1password.

Arguments: `[github username]`

### sg analytics reset

Delete all analytics stored locally.


### sg analytics view

View all analytics stored locally.


Flags:

* `--raw`: view raw data

## sg help

Get help and docs about sg.


Flags:

* `--full, -f`: generate full markdown sg reference
* `--help, -h`: show help
* `--output="<value>"`: write reference to `file`

## sg version

View details for this installation of sg.


### sg version changelog

See what's changed in or since this version of sg.


Flags:

* `--limit="<value>"`: Number of changelog entries to show. (default: 5)
* `--next`: Show changelog for changes you would get if you upgrade.

## sg update

Update local sg installation.

Update local sg installation with the latest changes. To see what's new, run:

    sg version changelog -next


## sg logo

Print the sg logo.

By default, prints the sg logo in different colors. When the 'classic' argument is passed it prints the classic logo.

Arguments: `[classic]`
