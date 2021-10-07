# sg - the Sourcegraph developer tool

```
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

`sg` is the CLI tool that Sourcegraph developers can use to develop Sourcegraph.

- [Quickstart](#quickstart)
- [Installation](#quickstart)
- [Usage](#usage)
  - [`sg [start,run,run-set]` - Start dev environment](#sg-startrunrun-set---start-dev-environment)
  - [`sg test` - Running test suites](#sg-test---running-test-suites)
  - [`sg doctor` - Check health of dev environment](#sg-doctor---check-health-of-dev-environment)
  - [`sg live` - See currently deployed version](#sg-live---see-currently-deployed-version)
  - [`sg migration` - Run or manipulate database migrations](#sg-migration---run-or-manipulate-database-migrations)
  - [`sg rfc` - List, open, or search Sourcegraph RFCs](#sg-rfc---list-or-open-sourcegraph-rfcs)
  - [`sg teammate [time,handbook]` - Get current time, or open their handbook page](#sg-rfc---get-current-time-or-open-their-handbook-page)
  - [`sg ci [preview]` - Preview build steps of the current branch on the CI](#sg-rfc---preview-build-steps-of-the-current-branch-on-the-ci)
    list-or-open-sourcegraph-rfcs)
- [Configuration](#configuration)
- [Contributing to sg](#contributing-to-sg)

## Quickstart

1. Install the [Sourcegraph development dependencies](https://docs.sourcegraph.com/dev/getting-started/quickstart_1_install_dependencies).
2. In your clone of [`sourcegraph/sourcegraph`](https://github.com/sourcegraph/sourcegraph), run:

   ```
   ./dev/sg/install.sh
   ```

3. Start the default Sourcegraph environment:

   ```
   sg start
   ```

   Once the `web` process has finished compilation, open [`https://sourcegraph.test:3443`](https://sourcegraph.test:3443/) in your browser.

## Installation

**`sg` requires the [Sourcegraph development dependencies](https://docs.sourcegraph.com/dev/getting-started/quickstart_1_install_dependencies) to be installed.**

### Using install script (recommended)

Run the following in the root of `sourcegraph/sourcegraph`:

```
./dev/sg/install.sh
```

That builds the `sg` binary and moves it to the standard installation location for Go binaries.

If you don't have a `$GOPATH` set (or don't know what that is), that location is `$HOME/go/bin`. If you do use `$GOPATH` the location is `$GOPATH/bin`.

Make sure that location is in your `$PATH`. (If you use `$GOPATH` then `$GOPATH/bin` needs to be in the `$PATH`)

> **Note for Linux users:** A command called [sg](https://www.man7.org/linux/man-pages/man1/sg.1.html) is already available at `/usr/bin/sg`. To use the Sourcegraph `sg` CLI, you need to make sure that its location comes first in `PATH`. For example, by prepending `$GOPATH/bin`:
>
> ```
> export PATH=$GOPATH/bin:$PATH
> ```
>
> Instead of the more conventional:
>
> ```
> export PATH=$PATH:$GOPATH/bin
> ```
>
> Or you may add an alias to your `.bashrc`:
>
> ```
> alias sg=$HOME/go/bin/sg
> ```

### Manually building the binary

If you want full control over where the `sg` binary ends up, use this option.

In the root of `sourcegraph/sourcegraph`, run:

```
go build -o ~/my/path/sg ./dev/sg
```

Then make sure that `~/my/path` is in your `$PATH`.

## Usage

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
sg live dot-com
sg live k8s

# See which version is deployed on a custom environment
sg live https://demo.sourcegraph.com

# List environments:
sg live -help
```

### `sg migration` - Run or manipulate database migrations

```bash
# Migrate local default database up
sg migration up

# Migrate specific database down one migration
sg migration down --db codeintel -n 1

# Add new migration for specific database
sg migration add --db codeintel 'add missing index'

# Squash migrations for default database
sg migration squash

# Fixup your migrations comapred to main for databases
sg migration fixup

# To see what operations `sg migration fixup` will run, you can check with
sg migration fixup -run=false

# Or to run for only one database, you can use the -db flag, as in other operations.
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

### `sg ci` - Preview build steps of the current branch on the CI

```bash
# Preview build steps and displays detected changes
# (must be a on branch and accuracy depends on origin/main)
sg ci preview
```

## Configuration

`sg` is configured through the [`sg.config.yaml` file in the root of the `sourcegraph/sourcegraph` repository](https://github.com/sourcegraph/sourcegraph/blob/main/sg.config.yaml). Take a look at that file to see which commands are run in which environment, how these commands set setup, what environment variables they use, and more.

To modify your configuration locally, you can overwrite chunks of configuration by creating a `sg.config.overwrite.yaml` file in the root of the repository. It's `.gitignore`d so you won't accidentally commit those changes.

If an `sg.config.overwrite.yaml` file exists, its contents will be merged with the content of `sg.config.yaml`, overwriting where there are conflicts. This is useful for running custom command sets or adding environment variables
specific to your work.

You can run `sg run debug-env` to see the environment variables passed `sg`'s child processes.

### Examples

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

## Contributing to `sg`

Want to hack on `sg`? Great! Here's how:

1. Read through the [`sg` Vision](./VISION.md) to get an idea of what `sg` should be in the long term.
2. Look at the open [`sg` issues](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+label%3Asg)

When you want to hack on `sg` it's best to be in the `dev/sg` directory and run it from there:

```
cd dev/sg
go run . -config ../../sg.config.yaml start
```

The `-config` can be anything you want, of course.
