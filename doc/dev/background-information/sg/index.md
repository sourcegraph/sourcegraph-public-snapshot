# `sg` - the Sourcegraph developer tool

```none
          _____                    _____
         /\    \                  /\    \
        /::\    \                /::\    \
       /::::\    \              /::::\    \
      /::::::\    \            /::::::\    \
     /:::/\:::\    \          /:::/\:::\    \
    /:::/__\:::\    \        /:::/  \:::\    \
    \:::\   \:::\    \      /:::/    \:::\    \
  ___\:::\   \:::\    \    /:::/    / \:::\    \
 /\   \:::\   \:::\    \  /:::/    /   \:::\ ___\
/::\   \:::\   \:::\____\/:::/____/  ___\:::|    |
\:::\   \:::\   \::/    /\:::\    \ /\  /:::|____|
 \:::\   \:::\   \/____/  \:::\    /::\ \::/    /
  \:::\   \:::\    \       \:::\   \:::\ \/____/
   \:::\   \:::\____\       \:::\   \:::\____\
    \:::\  /:::/    /        \:::\  /:::/    /
     \:::\/:::/    /          \:::\/:::/    /
      \::::::/    /            \::::::/    /
       \::::/    /              \::::/    /
        \::/    /                \::/____/
         \/____/

```

[`sg`](https://github.com/sourcegraph/sourcegraph/tree/main/dev/sg) is the CLI tool that Sourcegraph developers can use to develop Sourcegraph.
Learn more about the tool's overall vision in [`sg` Vision](./vision.md), and how to use it in the [usage section](#usage).

> NOTE: Have feedback or ideas? Feel free to [open a discussion](https://github.com/sourcegraph/sourcegraph/discussions/categories/developer-experience)! Sourcegraph teammates can also leave a message in [#dev-experience](https://sourcegraph.slack.com/archives/C01N83PS4TU).

## Quickstart

1. Run the following to download and install `sg`:

   ```sh
   curl --proto '=https' --tlsv1.2 -sSLf https://install.sg.dev | sh
   ```

2. In your clone of [`sourcegraph/sourcegraph`](https://github.com/sourcegraph/sourcegraph), start the default Sourcegraph environment:

   ```sh
   sg start
   ```

3. Once the `enterprise-web` process has finished compilation, open [`https://sourcegraph.test:3443`](https://sourcegraph.test:3443/) in your browser.

A more detailed introduction is available in the [development quickstart guide](../../setup/quickstart.md).

## Installation

### Using pre-built binaries (recommended)

Run the following command in a terminal:

```sh
curl --proto '=https' --tlsv1.2 -sSLf https://install.sg.dev | sh
```

That will download the latest release of `sg` from [here](https://github.com/sourcegraph/sg/releases), put it in a temporary location and run `sg install` to install it to a permanent location in your `$PATH`.

### Manually building the binary

> NOTE: **This method requires that Go has already been installed according to the [development quickstart guide](../../setup/quickstart.md).**

If you want full control over where the `sg` binary ends up, use this option.

In the root of `sourcegraph/sourcegraph`, run:

```sh
go build -o ~/my/path/sg ./dev/sg
```

Then make sure that `~/my/path` is in your `$PATH`.

> NOTE: **For Linux users:** A command called [sg](https://www.man7.org/linux/man-pages/man1/sg.1.html) is already available at `/usr/bin/sg`. To use the Sourcegraph `sg` CLI, you need to make sure that its location comes first in `PATH`. For example, by prepending `$GOPATH/bin`:
>
> `export PATH=$GOPATH/bin:$PATH`
>
> Instead of the more conventional:
>
> `export PATH=$PATH:$GOPATH/bin`
>
> Or you may add an alias to your `.bashrc`:
>
> `alias sg=$HOME/go/bin/sg`

## Updates

Once set up, `sg` will automatically check for updates and update itself if a change is detected in your local copy of `origin/main`.

To force a manual update of `sg`, run:

```sh
sg update
```

In order to temporarily turn off automatic updates, run your commands with the `-skip-auto-update` flag: 

```sh
sg -skip-auto-update [cmds ...]
```

On the next command run, if a new version is detected, `sg` will auto update before running.

> NOTE: This feature requires that Go has already been installed according to the [development quickstart guide](../../setup/quickstart.md).

## Usage

See [configuration](#configuration) to learn more about configuring `sg` behaviour.

### Help

You can get help about commands in a variety of ways:

```sh
sg help # show all available commands

# learn about a specific command or subcommand
sg <command> -h
sg <command> --help
```

### Autocompletion

If you have used `sg setup`, you should have autocompletions set up for `sg`. To enable it, type out a partial command and press the <kbd>Tab</kbd> key twice. For example:

```none
sg start<tab><tab>
```

To get autocompletions for the available flags for a command, type out a command and `-` and press the <kbd>Tab</kbd> key twice. For example:

```none
sg start -<tab><tab>
```

Both of the above work if you provide partial values as well to narrow down the suggestions. For example, the following will suggest run sets that start with `web-`:

```none
sg start web-<tab><tab>
```

### `sg start` - Start dev environments

```bash
# Run default environment, Sourcegraph enterprise:
sg start

# List available environments (defined under `commandSets` in `sg.config.yaml`):
sg start -help

# Run the enterprise environment with code-intel enabled:
sg start enterprise-codeintel

# Run the environment for Batch Changes development:
sg start batches

# Override the logger levels for specific services
sg start --debug=gitserver --error=enterprise-worker,enterprise-frontend enterprise
```

### `sg run` - Run single commands

```bash
# Run specific commands:
sg run gitserver
sg run frontend

# List available commands (defined under `commands:` in `sg.config.yaml`):
sg run -help

# Run multiple commands:
sg run gitserver frontend repo-updater
```

### `sg test` - Running test suites

```bash
# Run different test suites:
sg test backend
sg test backend-integration
sg test frontend
sg test frontend-e2e

# List available test suites:
sg test -help

# Arguments are passed along to the command
sg test backend-integration -run TestSearch
```

### `sg doctor` - Check health of dev environment

```bash
# Run the checks defined in sg.config.yaml
sg doctor
```

### `sg live` - See currently deployed version

```bash
# See which version is deployed on a preset environment
sg live cloud
sg live k8s

# See which version is deployed on a custom environment
sg live https://demo.sourcegraph.com

# List environments:
sg live -help
```

### `sg migration` - Run or manipulate database migrations

```bash
# Migrate local default database up all the way
sg migration up

# Migrate specific database down one migration
sg migration down --db codeintel

# Add new migration for specific database
sg migration add --db codeintel 'add missing index'

# Squash migrations for default database
sg migration squash
```

### `sg rfc` - List or open Sourcegraph RFCs

```bash
# List all RFCs
sg rfc list

# Search for an RFC
sg rfc search "search terms"

# Open a specific RFC
sg rfc open 420
```

### `sg ci` - Interact with Sourcegraph's continuous integration

Interact with Sourcegraph's [continuous integration](https://docs.sourcegraph.com/dev/background-information/ci) pipelines on [Buildkite](https://buildkite.com/sourcegraph).

```bash
# Preview what a CI run for your current changes will look like
sg ci preview

# Check on the status of your changes on the current branch in the Buildkite pipeline
sg ci status
# Check on the status of a specific branch instead
sg ci status --branch my-branch
# Block until the build has completed (it will send a system notification)
sg ci status --wait
# Get status for a specific build number
sg ci status --build 123456 

# Pull logs of failed jobs to stdout
sg ci logs
# Push logs of most recent main failure to local Loki for analysis
# You can spin up a Loki instance with 'sg run loki grafana'
sg ci logs --branch main --out http://127.0.0.1:3100
# Get the logs for a specific build number, useful when debugging
sg ci logs --build 123456 

# Manually trigger a build on the CI with the current branch
sg ci build 
# Manually trigger a build on the CI on the current branch, but with a specific commit
sg ci build --commit my-commit
# Manually trigger a main-dry-run build of the HEAD commit on the current branch
sg ci build main-dry-run
sg ci build --force main-dry-run
# Manually trigger a main-dry-run build of a specified commit on the current ranch
sg ci build --force --commit my-commit main-dry-run
# View the available special build types
sg ci build --help
```

### `sg teammate` - Get current time or open their handbook page

```bash
# Get the current time of a team mate based on their slack handle (case insensitive).
sg teammate time @dax
sg teammate time dax
# or their full name (case insensitive)
sg teammate time thorsten ball

# Open their handbook bio
sg teammate handbook asdine
```

### `sg secret` - Interact with `sg` secrets

```bash
# List all secrets stored in your local configuration. 
sg secret list

# Remove the secrets associated with buildkite (sg ci build)
sg secret reset buildkite

```

### `sg lint` - Run linters against local code

```bash
# Run all possible checks 
sg lint

# Run only go related checks
sg lint go

# Run only shell related checks
sg lint shell

# Run only client related checks
sg lint client

# List all available check groups 
sg lint --help
```

### `sg db` - Interact with your local Sourcegraph database(s)

```bash
# Reset the Sourcegraph 'frontend' database
sg db reset-pg

# Reset the 'frontend' and 'codeintel' databases
sg db reset-pg -db=frontend,codeintel

# Reset all databases ('frontend', 'codeintel', 'codeinsights')
sg db reset-pg -db=all

# Reset the redis database
sg db reset-redis

# Create a site-admin user whose email and password are foo@sourcegraph.com and sourcegraph.
sg db add-user -name=foo
```

### `sg update` - Update sg itself

```bash 
# Manually update sg
sg update
```

## Configuration

Default `sg` behaviour is configured through the [`sg.config.yaml` file in the root of the `sourcegraph/sourcegraph` repository](https://github.com/sourcegraph/sourcegraph/blob/main/sg.config.yaml). Take a look at that file to see which commands are run in which environment, how these commands set setup, what environment variables they use, and more.

**To modify your configuration locally, you can overwrite chunks of configuration by creating a `sg.config.overwrite.yaml` file in the root of the repository.** It's `.gitignore`d so you won't accidentally commit those changes.

If an `sg.config.overwrite.yaml` file exists, its contents will be merged with the content of `sg.config.yaml`, overwriting where there are conflicts. This is useful for running custom command sets or adding environment variables
specific to your work.

You can run `sg run debug-env` to see the environment variables passed `sg`'s child processes.

### Configuration examples

#### Changing database configuration

In order to change the default database configuration, the username and the database, for example, create an `sg.config.overwrite.yaml` file that looks like this:

```yaml
env:
  PGUSER: 'mrnugget'
  PGDATABASE: 'my-database'
```

That works for all the other `env` variables in `sg.config.yaml` too.

#### Defining a custom environment by setting a `commandset`

You can customize what boots up in your development environment by defining a `commandSet` in your `sg.config.overwrite.yaml`.

For example, the following defines a commandset called `minimal-batches` that boots up a minimal environment to work on Batch Changes:

```yaml
commandsets:
  minimal-batches:
    checks:
      - docker
      - redis
      - postgres
    commands:
      - enterprise-frontend
      - enterprise-worker
      - enterprise-repo-updater
      - enterprise-web
      - gitserver
      - searcher
      - symbols
      - caddy
      - github-proxy
      - zoekt-indexserver-0
      - zoekt-indexserver-1
      - zoekt-webserver-0
      - zoekt-webserver-1
      - batches-executor-firecracker
```

With that in `sg.config.overwrite.yaml` you can now run `sg start minimal-batches`.

### Attach a debugger

To attach the [Delve](https://github.com/go-delve/delve) debugger, pass the environment variable `DELVE=true` into `sg`. [Read more here](https://docs.sourcegraph.com/dev/how-to/debug_live_code#debug-go-code)

### Offline development

Sometimes you will want to develop Sourcegraph but it just so happens you will be on a plane or a
train or perhaps a beach, and you will have no WiFi. And you may raise your fist toward heaven and
say something like, "Why, we can put a man on the moon, so why can't we develop high-quality code
search without an Internet connection?" But lower your hand back to your keyboard and fret no
further, you *can* develop Sourcegraph with no connectivity by setting the
`OFFLINE` environment variable:

```bash
OFFLINE=true sg start
```

## Contributing to `sg`

Want to hack on `sg`? Great! Here's how:

1. Read through the [`sg` Vision](./vision.md) to get an idea of what `sg` should be in the long term.
2. Explore the [`sg` source code](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/dev/sg).
3. Look at the open [`sg` issues](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+label%3Asg).

When you want to hack on `sg` it's best to be in the `dev/sg` directory and run it from there:

```sh
cd dev/sg
go run . -config ../../sg.config.yaml start
```

The `-config` can be anything you want, of course.

Have questions or need help? Feel free to [open a discussion](https://github.com/sourcegraph/sourcegraph/discussions/categories/developer-experience)! Sourcegraph teammates can also leave a message in [#dev-experience](https://sourcegraph.slack.com/archives/C01N83PS4TU).

> NOTE: For Sourcegraph teammates, we have a weekly [`sg` hack hour](https://handbook.sourcegraph.com/departments/product-engineering/engineering/enablement/dev-experience#sg-hack-hour) you can hop in to if you're interested in contributing!

## Dockerized sg

A `sourcegraph/sg` Docker image is available: 

```
# ... 
COPY --from us.gcr.io/sourcegraph-dev/sg:insiders /usr/local/bin/sg ./sg
# ...
```

### Development tips

- Due to [#29222](https://github.com/sourcegraph/sourcegraph/issues/29222), you might need to set `CONFIGURATION_MODE: 'empty'` if you encounter errors where `sg` tries to connect to `frontend`.
