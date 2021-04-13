# sg - the Sourcegraph developer tool

`sg` is the CLI tool that Sourcegraph developers can use to develop Sourcegraph.

- [Quickstart](#quickstart)
- [Installation](#installation)
- [Usage](#usage)
  - [Start dev environment](#start-dev-environment)
  - [Running tests](#running-tests)
  - [Generators](#generators)
  - [Database migrations](#database-migrations)
  - [Edit configuration files](#edit-configuration-files)
- [Configuration](#configuration)
- [TODOs](#todos)
- [Hacking](#hacking)

## Quickstart

Run the following to install `sg` from the `main` branch:

```
go install github.com/sourcegraph/sourcegraph/dev/sg@latest
```

Make sure that `$HOME/go/bin` is in your `$PATH`. (If you use `$GOPATH` then `$GOPATH/bin` needs to be in the `$PATH`)

**Note for Linux users:** A command called [sg](https://www.man7.org/linux/man-pages/man1/sg.1.html) is already available at `/usr/bin/sg`. To use the Sourcegraph `sg` CLI, you need to make sure that its location comes first in `PATH`. For example, by prepending `$GOPATH/bin`:

```
export PATH=$GOPATH/bin:$PATH
```

Instead of the more conventional:

```
export PATH=$PATH:$GOPATH/bin
```

Or you may add an alias to your `.bashrc`:

```
alias sg=$HOME/go/bin/sg
```

Then, in the root of `sourcegraph/sourcegraph`, run:

```
sg start
```

This will boot the `default` commands in `sg.config.yaml` in the root of the repository.

**Alternative install method** (if you want to move the binary to a custom location):

In the root of `sourcegraph/sourcegraph`, run the following:

```
go build -o ~/my/path/sg ./dev/sg
```

Make sure that `~/my/path` is in your `$PATH` then.

## Inspiration

- [GitLab Developer Kit (GDK)](https://gitlab.com/gitlab-org/gitlab-development-kit)
- Stripe's `pay` command, [described here](https://buttondown.email/nelhage/archive/papers-i-love-gg/)
- [Stack Exchange Local Environment Setup](https://twitter.com/nick_craver/status/1375871107773956103?s=21) command

## Usage

### Start dev environment

```bash
# Run default environment (this starts the 'default' command set defined in config):
sg start

# Run the enterprise environment:
sg run-set enterprise

# Run specific commands:
sg run gitserver
sg run frontend

# Run predefined sets of commands:
sg run-set enterprise

# TODO: Rebuild and restart a command (if it has `build` defined, see Configuration)
sg build gitserver
```

### Running tests

```bash
# Run go tests
sg test backend

# Run web e2e integration tests
sg test frontend-e2e

# Run other tests
sg test backend-integration
sg test frontend-snapshot
sg test regression

# Without argument it lists all available tests:
sg test

# TODO: Arguments are passed along to the command
sg test backend-integration -run TestSearch
```

### Generators

TODO: Build this

```bash
# Generate code
sg generate

# Generate only specific things
sg generate mocks ./internal/enterprise/batches
```

### Database migrations

TODO: Build this

```bash
# Create a new migration
sg migration new --name=my-new-migration
# Create a new migration for the codeintel database
sg migration new --db=codeintel --name=my-new-migration

# Run all migrations _up_
sg migration up
sg migration down
```

### Edit configuration files

TODO: Build this

```bash
# Edit the site configuration
sg edit site-config # opens site-config in $EDITOR
# Edit external service configuration
sg edit external-services # opens external-services.json in $EDITOR
```

### Tail logs

TODO: Build this

```bash
# Tail the SQL logs
sg tail-log sql
# Tail the http logs
sg tail-log http
# Tail all logs
sg tail-log all
```

## Configuration

`sg` is configured through the `sg.yaml` file in the root of the `sourcegraph/sourcegraph` repository:

```yaml
commands:
  gitserver:
    cmd: .bin/gitserver
    install: go install github.com/sourcegraph/sourcegraph/cmd/gitserver -o .bin/gitserver
    env:
      HOSTNAME: $SRC_GIT_SERVER_1

  enterprise-gitserver:
    cmd: .bin/gitserver
    install: go install github.com/sourcegraph/sourcegraph/enterprise/cmd/gitserver -o .bin/gitserver
    env:
      HOSTNAME: $SRC_GIT_SERVER_1

  query-runner:
    cmd: .bin/query-runner
    install: go install github.com/sourcegraph/sourcegraph/cmd/query-runner -o .bin/gitserver

  searcher:
    cmd: .bin/searcher
    install: go install github.com/sourcegraph/sourcegraph/cmd/searcher -o .bin/gitserver

  caddy:
    install_doc.darwin: 'use brew install'
    install_doc.linux: 'use apt install'

  web:
    cmd: ./node_modules/.bin/gulp --silent --color dev
    install: yarn install

  # [...]

commandsets:
  # This is the set that will be used when `sg start` is run:
  default:
    - gitserver
    - query-runner
    - repo-updater
    - searcher
    - symbols
    - github-proxy
    - frontend
    - watch
    - caddy
    - web
    - syntect_server
    - zoekt-indexserver-0
    - zoekt-indexserver-1
    - zoekt-webserver-0
    - zoekt-webserver-1
  # Another set that can be run with `sg run-set monitoring`:
  monitoring:
    - jaeger
    - docsite
    - prometheus
    - grafana
    - postgres_exporter
  enterprise:
    - enterprise-gitserver
    - enterprise-query-runner
    - enterprise-repo-updater
    - enterprise-frontend
    - searcher
    - symbols
    - github-proxy
    - watch
    - caddy
    - web
    - syntect_server
    - zoekt-indexserver-0
    - zoekt-indexserver-1
    - zoekt-webserver-0
    - zoekt-webserver-1

tests:
  # These can be run with `sg test [name]`
  backend:
    cmd: go test ./...
  backend-integration:
    cmd: go test -long -base-url $BASE_URL -email $EMAIL -username $USERNAME -password $PASSWORD
    env:
      # These are defaults. They can be overwritten by setting the env vars when
      # running the command.
      BASE_URL: 'http://localhost:3080'
      EMAIL: 'joe@sourcegraph.com'
      PASSWORD: '12345'
  frontend:
    cmd: yarn run jest --testPathIgnorePatterns end-to-end regression integration storybook
  frontend-e2e:
    cmd: yarn run mocha ./client/web/src/end-to-end/end-to-end.test.ts
    env:
      TS_NODE_PROJECT: client/web/src/end-to-end/tsconfig.json
```

## TODOs

- [ ] All of the things marked as TODOs above
- [ ] Add the remaining processes from `<root>/dev/Procfile` to `<root>/sg.config.yaml`
- [ ] Add a _simple_ way to define in the config file when a restart after a rebuild is not necessary
  - Something like `check_binary: .bin/frontend` which would take a SHA256 before and after rebuild and only restart if SHA doesn't match
- [ ] Rename `install` to `build` because it's clearer
- [ ] Add support for "dev environment setup"
  - Something like `sg check` which runs `check_cmds` in the config file and provides helpful output if one of them failed ("check_cmd postgres failed. Install postgres with...")
- [ ] Add built-in support for "download binary" so that the `caddy` command, for example, would be 3 lines instead of 20

## Hacking

When you want to hack on `sg` it's best to be in the `dev/sg` directory and run
it from there:

```
cd dev/sg
go run . -config ../../sg.config.yaml start
```

The `-config` can be anything you want, of course.
