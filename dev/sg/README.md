# sg - the Sourcegraph developer tool

`sg` is the CLI tool that Sourcegraph developers can use to develop Sourcegraph.

- [TODOs](#todos)
- [Installation](#installation)
- [Usage](#usage)
  - [Start dev environment](#start-dev-environment)
  - [Running tests](#running-tests)
  - [Generators](#generators)
  - [Database migrations](#database-migrations)
  - [Edit configuration files](#edit-configuration-files)
- [Configuration](#configuration)

## Quickstart

This is mostly meant for developing `sg`:

```
go run . -config sg.config.example.yaml start
```

This compiles and starts `sg`, starting the `default` command set defined in `sg.config.example.yaml`, which boots up our dev environment.

## TODOs

- [ ] Build everything below

## Installation

```
go get github.com/sourcegraph/sourcegraph/dev/sg
```

## Usage

### Start dev environment

```bash
# Run complete environment:
sg start

# Run specific commands:
sg run gitserver
sg run frontend

# Run predefined sets of commands:
sg run-set backend
sg run-set monitoring

# Rebuild and restart a command (if it has `build` defined, see Configuration)
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

# Arguments are passed along to the command
sg test backend-integration -run TestSearch
```

### Generators

```bash
# Generate code
sg generate

# Generate only specific things
sg generate mocks ./internal/enterprise/batches
```

### Database migrations

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

```bash
# Edit the site configuration
sg edit site-config # opens site-config in $EDITOR
# Edit external service configuration
sg edit external-services # opens external-services.json in $EDITOR
```

### Tail logs

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
