# Auto-indexing configuration reference

<aside class="beta">
<p>
<span class="badge badge-beta">Beta</span> This feature is in beta for self-hosted customers and might change in the future.
</p>

<p><b>We're very much looking for input and feedback on this feature.</b> You can either <a href="https://about.sourcegraph.com/contact">contact us directly</a>, <a href="https://github.com/sourcegraph/sourcegraph">file an issue</a>, or <a href="https://twitter.com/sourcegraph">tweet at us</a>.</p>
</aside>

This document details the expected contents of [explicit index job configuration](../how-to/configure_auto_indexing.md#explicit-index-job-configuration) for a particular repository. Depending on how this configuration is supplied, it may be encoded as JSON.

## Keys

The root of the configuration has two top-level keys, `index_jobs` and `shared_steps`, documented below.

### [`index_jobs`](#index-jobs)

The index jobs field defines a set of [index job objects](#index-job-object) describing the actions to perform to successfully index a fresh clone of a repository. Each index job is executed independently (and possibly in parallel) by an executor.

### [`shared_steps`](#shared-steps)

The shared steps field defines an ordered sequence of pre-indexing actions (formatted as a [Docker step object](#docker-step-object)). Each step is executed before each index job is executed. Shared steps are executed individually once _per indexing job_ and are designed to be used as a way of sharing code and common configuration for projects within a repository (but are not designed as a way to share compute resources).

## Examples

In the following example, we have a repository configured to index three different projects: `cmd/foo`, `cmd/bar`, and `cmd/baz`. These projects are executed concurrently by available executor processes. Each index job shares an initial step that runs a `setup.sh` script in the root of the repository and installs Go dependencies. The index job for `cmd/bar` additionally requires a code generation step, which is performed after the shared step but before the indexer invocation.

```yaml
shared_steps:
  image: sourcegraph/lsif-go
  commands:
    - ./setup.sh      # Run a custom setup script
    - go mod download # Download dependencies

index_jobs:
  # Index cmd/foo
  - indexer: sourcegraph/lsif-go
    root: cmd/foo

  # Index cmd/bar
  # Requires an additional codegen step
  - indexer: sourcegraph/lsif-go
    steps:
      - image: sourcegraph/lsif-go
        commands:
          - go generate ./...
        root: cmd/bar
    root: cmd/bar
  
  # Index cmd/baz
  - indexer: sourcegraph/lsif-go
    root: cmd/baz
```

## Index job object

Each configured index job is run by a single executor process. Multiple index jobs configured for the same repository may be executed by different executor instances, out of order, and possibly in parallel.

The basic outline of an index job is as follows.

1. The target repository is cloned into the job's workspace from Sourcegraph.
2. Pre-indexing steps (if any are configured) are executed sequentially in individual Docker containers that volume mount the workspace.
3. A separate Docker container for the indexer is started.
4. Local steps (if configured) are executed within the running container.
5. The indexer is invoked within the running container to produce an index artifact.
6. The [`src` CLI](../../cli/index.md) binary is invoked to upload the index artifact to the Sourcegraph instance.

The pre-indexing steps, indexer container, local steps, and indexer arguments are configurable via this object.

### Keys

Each indexing job object can be configured with the following keys.

#### [`steps`](#index-job-steps)

The steps field defines an ordered sequence of pre-indexing actions (formatted as a [Docker step object](#docker-step-object)). If `shared_steps` are defined in the top level of the configuration, those steps are **prepended** to this sequence. Each step is executed before the indexer itself is invoked.

#### [`indexer`](#index-job-indexer)

The name of the Docker image distribution of the target indexer.

#### [`local_steps`](#index-job-local-steps)

An ordered sequence of commands to execute within a container running the configured indexer image. These commands are passed directly into the target container via `docker exec`, one at a time. Local steps should be used over Docker steps when the intended side effects alter state outside of the workspace on disk (e.g., setting up environment variables or installing OS packages require by the indexer tool).

#### [`indexer_args`](#index-job-indexer-args)

An ordered sequence of arguments that make up the indexing command. The indexing command is passed directly into the target container via `docker exec`. This step is expected to produce a code graph index artifact (as described by the `root` and `outfile` fields, described below).

#### [`root`](#index-job-root)

The working directory within the Docker container where the provided local steps and indexer commands are executed. This working directory is relative to the root of the target repository. An empty value (the default) indicates the root of the repository. This is also the directory relative to the path where the code graph index artifact is produced.

#### [`outfile`](#index-job-outfile)

The path to the code graph index artifact produced by the indexer, which is uploaded to the target Sourcegraph instance via [`src` CLI](../../cli/index.md) after the index step has completed successfully. This path is relative to the index job `root` (defined above). If not supplied, the value is assumed to be `dump.lsif` (which is the default artifact name of many indexers).

Supply this argument when the target indexer produces a differently named artifact. Alternatively, some indexers provide flags to change the artifact name; in which case `dump.lsif` can be supplied there and a value for this key can be omitted.

### Examples

The following example uses the Docker image `sourcegraph/lsif-go` pinned at the tag `v1.6.7` and additionally secured with an image digest. This index configuration runs the Go indexer with quiet output in the `dev/sg` directory and uploads the resulting index file (`dump.lsif` by default).

```yaml
indexer: sourcegraph/lsif-go:v1.6.7@sha256::9c2d9cf...1baed2b

# sourcegraph/lsif-go does not define a Docker entrypoint, so we need to
# indicate which binary to invoke (along with any additional arguments).
# Here, -q and --no-animation are lsif-go flags that control verbosity.
indexer_args:
  - lsif-go
  - -q
  - --no-animation

root: dev/sg
```

The following example uses the Docker image `sourcegraph/scip-typescript` pinned at the tag `autoindex`. This index configuration will run `npm install` followed by the TypeScript indexer `scip-typescript` in the `editors/code` directory. In this job, both commands run in the same workspace, but are invoked in different Docker containers. Both containers happen to be based on the same image, but that's not necessary.

```yaml
steps:
  - image: sourcegraph/scip-typescript:autoindex
    commands:
      - npm install
    root: editors/code

# sourcegraph/scip-typescript does not define a Docker entrypoint, so we need
# to indicate which binary to invoke (along with any additional arguments).
indexer: sourcegraph/scip-typescript:autoiondex
indexer_args:
    - scip-typescript
    - index

root: editors/code
```

The following example uses the Docker image `sourcegraph/lsif-clang` using the default tag. This index configuration will run a series of commands followed by the Clang indexer. Because these commands are defined as _local_ steps, they will occur in the same Docker container as the index operation. This distinction is necessary when the setup process must affect more than the filesystem - here we needed to install OS packages and install a build-time tool.

```yaml
indexer: sourcegraph/lsif-clang
local_steps:
  - apt-get update
  - DEBIAN_FRONTEND=noninteractive apt-get -y install build-essential ... clang-10
  - python3 -m pip install compiledb
  - python3 -m compiledb -n make

# sourcegraph/lsif-clang does not define a Docker entrypoint, so we need 
# to indicate which binary to invoke (along with any additional arguments).
# Here, we have lsif-clang consume the compile_commands.json manifest generated
# by the local steps defined above.
indexer_args:
  - lsif-clang
  - compile_commands.json

# Ensure lsif-clang generates the file expected by src code-intel upload.
outfile: dump.lsif
```

## Docker step object

Each configured Docker step is executed sequentially using the same volume-mounted workspace, which is initially seeded with a fresh clone of the target repository. If the contents of the workspace is modified by one step, the changes will be visible in the next. This makes Docker steps especially useful for pre-indexing tasks that require modification of the source code, such as dependency resolution (e.g., `npm install`, `go mod download`) and code generation (e.g., `go generate ./...`) required for successful compilation or indexing of the project.

### Keys

Each Docker step object can be configured with the following keys.

#### [`image`](#docker-step-image)

The name of the Docker image in which the configured commands are executed.

#### [`commands`](#docker-step-commands)

An ordered sequence of commands to execute within a container running the configured image. These commands are passed directly into the target container via `docker exec`, one at a time.

#### [`root`](#docker-step-root)

The working directory within the Docker container where the provided commands are executed. This working directory is relative to the root of the target repository. An empty value (the default) indicates the root of the repository.

### Examples

The following example runs `go generate` over all packages in the root of the repository.

```yaml
indexer: golang:1.17.3-buster@sha256:c1bae5f...3e4f07b
commands:
  - go generate ./...
```

The following example runs a series of commands to install dependencies in the `lib/proj` directory.

```yaml
indexer: node:alpine3.12@sha256:4435145...fd06fd3
commands:
  - yarn config set ignore-engines true
  - yarn install
root: lib/proj
```
