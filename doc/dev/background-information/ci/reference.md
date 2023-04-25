<!-- DO NOT EDIT: generated via: go generate ./enterprise/dev/ci -->

# Pipeline types reference

This is a reference outlining what CI pipelines we generate under different conditions.

To preview the pipeline for your branch, use `sg ci preview`.

For a higher-level overview, please refer to the [continuous integration docs](https://docs.sourcegraph.com/dev/background-information/ci).

## Run types

### Pull request

The default run type.

- Pipeline for `Go` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
  - **Linters and static analysis**: Run sg lint
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `Client` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
  - **Linters and static analysis**: Run sg lint
  - **Client checks**: Upload Storybook to Chromatic, Enterprise build, Build (client/jetbrains), Tests for VS Code extension, Integration tests for the Cody VS Code extension, ESLint (all), ESLint (web), Stylelint (all)
  - **Pipeline setup**: Trigger async
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `GraphQL` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
  - **Client checks**: Upload Storybook to Chromatic, Enterprise build, Build (client/jetbrains), Tests for VS Code extension, Integration tests for the Cody VS Code extension, ESLint (all), ESLint (web), Stylelint (all)
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `DatabaseSchema` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `Docs` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
  - **Linters and static analysis**: Run sg lint
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `Dockerfiles` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
  - **Linters and static analysis**: Run sg lint
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `ExecutorVMImage` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `ExecutorDockerRegistryMirror` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `CIScripts` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
  - **CI script tests**: test-trace-command.sh
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `Terraform` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `SVG` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
  - **Linters and static analysis**: Run sg lint
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `Shell` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
  - **Linters and static analysis**: Run sg lint
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `DockerImages` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `WolfiPackages` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `WolfiBaseImages` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

- Pipeline for `Protobuf` changes:
  - **Metadata**: Pipeline metadata
  - **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
  - **Linters and static analysis**: Run sg lint
<<<<<<< HEAD
  - Build Docker images
  - Build Docker images
  - Upload build trace
=======
>>>>>>> main

### Bazel Exp Branch

The run type for branches matching `bzl/`.
You can create a build of this run type for your changes using:

```sh
sg ci build bzl
```

Base pipeline (more steps might be included based on branch changes):

- **Metadata**: Pipeline metadata
- **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests

### Wolfi Exp Branch

The run type for branches matching `wolfi/`.
You can create a build of this run type for your changes using:

```sh
sg ci build wolfi
```

Base pipeline (more steps might be included based on branch changes):

- **Metadata**: Pipeline metadata
- **Wolfi image builds**: Build Wolfi-based batcheshelper, Build Wolfi-based embeddings, Build Wolfi-based executor-kubernetes, Build Wolfi-based frontend, Build Wolfi-based github-proxy, Build Wolfi-based gitserver, Build Wolfi-based llm-proxy, Build Wolfi-based loadtest, Build Wolfi-based migrator, Build Wolfi-based precise-code-intel-worker, Build Wolfi-based repo-updater, Build Wolfi-based searcher, Build Wolfi-based server, Build Wolfi-based symbols, Build Wolfi-based worker, Build Wolfi-based blobstore, Build Wolfi-based cadvisor, Build Wolfi-based codeinsights-db, Build Wolfi-based codeintel-db, Build Wolfi-based indexed-searcher, Build Wolfi-based jaeger-agent, Build Wolfi-based jaeger-all-in-one, Build Wolfi-based node-exporter, Build Wolfi-based opentelemetry-collector, Build Wolfi-based postgres-12-alpine, Build Wolfi-based postgres_exporter, Build Wolfi-based prometheus, Build Wolfi-based prometheus-gcp, Build Wolfi-based redis-cache, Build Wolfi-based redis-store, Build Wolfi-based redis_exporter, Build Wolfi-based search-indexer, Build Wolfi-based sg, Build Wolfi-based syntax-highlighter

### Release branch nightly healthcheck build

The run type for environment including `{"RELEASE_NIGHTLY":"true"}`.

Base pipeline (more steps might be included based on branch changes):

- **Metadata**: Pipeline metadata
- Trigger 5.0 release branch healthcheck build
- Trigger 4.5 release branch healthcheck build

### Browser extension nightly release build

The run type for environment including `{"BEXT_NIGHTLY":"true"}`.

Base pipeline (more steps might be included based on branch changes):

- ESLint (all)
- ESLint (web)
- Stylelint (all)
- Test (client/browser)
- Puppeteer tests for chrome extension
- Test (all)
- E2E for chrome extension

### VS Code extension nightly release build

The run type for environment including `{"VSCE_NIGHTLY":"true"}`.

Base pipeline (more steps might be included based on branch changes):

- ESLint (all)
- ESLint (web)
- Stylelint (all)
- Tests for VS Code extension

### App release build

The run type for branches matching `app/release` (exact match).

Base pipeline (more steps might be included based on branch changes):

- App release

### App insiders build

The run type for branches matching `app/insiders` (exact match).

Base pipeline (more steps might be included based on branch changes):

- App release

### Tagged release

The run type for tags starting with `v`.

Base pipeline (more steps might be included based on branch changes):

- **Metadata**: Pipeline metadata
- **Pipeline setup**: Trigger async
- **Image builds**: Build symbols, Build server, Build batcheshelper, Build blobstore2, Build bundled-executor, Build dind, Build embeddings, Build executor, Build executor-kubernetes, Build executor-vm, Build jaeger-agent, Build jaeger-all-in-one, Build llm-proxy, Build sg, Build alpine-3.14, Build blobstore, Build cadvisor, Build codeinsights-db, Build codeintel-db, Build frontend, Build github-proxy, Build gitserver, Build grafana, Build indexed-searcher, Build migrator, Build node-exporter, Build opentelemetry-collector, Build postgres-12-alpine, Build postgres_exporter, Build precise-code-intel-worker, Build prometheus, Build prometheus-gcp, Build redis-cache, Build redis-store, Build redis_exporter, Build repo-updater, Build search-indexer, Build searcher, Build worker
- **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
- **Linters and static analysis**: Run sg lint
- **Client checks**: Upload Storybook to Chromatic, Enterprise build, Build (client/jetbrains), Tests for VS Code extension, Integration tests for the Cody VS Code extension, ESLint (all), ESLint (web), Stylelint (all)
- **CI script tests**: test-trace-command.sh
- **Integration tests**: Backend integration tests (gRPC), Backend integration tests, Code Intel QA
<<<<<<< HEAD
- **End-to-end tests**: Executors E2E, Sourcegraph E2E, Sourcegraph Upgrade
- **Publish images**: batcheshelper, blobstore2, bundled-executor, dind, embeddings, executor, executor-kubernetes, executor-vm, jaeger-agent, jaeger-all-in-one, llm-proxy, sg, symbols, server, alpine-3.14, blobstore, cadvisor, codeinsights-db, codeintel-db, frontend, github-proxy, gitserver, grafana, indexed-searcher, migrator, node-exporter, opentelemetry-collector, postgres-12-alpine, postgres_exporter, precise-code-intel-worker, prometheus, prometheus-gcp, redis-cache, redis-store, redis_exporter, repo-updater, search-indexer, searcher, worker, Publish executor image, Publish executor binary, Publish docker registry mirror image
- Upload build trace
=======
- **End-to-end tests**: Executors E2E, Sourcegraph E2E, Sourcegraph QA, Sourcegraph Cluster (deploy-sourcegraph) QA, Sourcegraph Upgrade
- **Publish images**: alpine-3.14, cadvisor, codeinsights-db, codeintel-db, frontend, github-proxy, gitserver, grafana, indexed-searcher, jaeger-agent, jaeger-all-in-one, blobstore, blobstore2, node-exporter, postgres-12-alpine, postgres_exporter, precise-code-intel-worker, prometheus, prometheus-gcp, redis-cache, redis-store, redis_exporter, repo-updater, search-indexer, searcher, symbols, syntax-highlighter, worker, migrator, executor, executor-kubernetes, executor-vm, batcheshelper, opentelemetry-collector, embeddings, dind, bundled-executor, server, sg, llm-proxy, Publish executor image, Publish executor binary, Publish docker registry mirror image
>>>>>>> main

### Release branch

The run type for branches matching `^[0-9]+\.[0-9]+$` (regexp match).

Base pipeline (more steps might be included based on branch changes):

- **Metadata**: Pipeline metadata
- **Pipeline setup**: Trigger async
- **Image builds**: Build symbols, Build server, Build batcheshelper, Build blobstore2, Build bundled-executor, Build dind, Build embeddings, Build executor, Build executor-kubernetes, Build executor-vm, Build jaeger-agent, Build jaeger-all-in-one, Build llm-proxy, Build sg, Build alpine-3.14, Build blobstore, Build cadvisor, Build codeinsights-db, Build codeintel-db, Build frontend, Build github-proxy, Build gitserver, Build grafana, Build indexed-searcher, Build migrator, Build node-exporter, Build opentelemetry-collector, Build postgres-12-alpine, Build postgres_exporter, Build precise-code-intel-worker, Build prometheus, Build prometheus-gcp, Build redis-cache, Build redis-store, Build redis_exporter, Build repo-updater, Build search-indexer, Build searcher, Build worker
- **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
- **Linters and static analysis**: Run sg lint
- **Client checks**: Upload Storybook to Chromatic, Enterprise build, Build (client/jetbrains), Tests for VS Code extension, Integration tests for the Cody VS Code extension, ESLint (all), ESLint (web), Stylelint (all)
- **CI script tests**: test-trace-command.sh
- **Integration tests**: Backend integration tests (gRPC), Backend integration tests, Code Intel QA
<<<<<<< HEAD
- **End-to-end tests**: Executors E2E, Sourcegraph E2E, Sourcegraph Upgrade
- **Publish images**: batcheshelper, blobstore2, bundled-executor, dind, embeddings, executor, executor-kubernetes, executor-vm, jaeger-agent, jaeger-all-in-one, llm-proxy, sg, symbols, server, alpine-3.14, blobstore, cadvisor, codeinsights-db, codeintel-db, frontend, github-proxy, gitserver, grafana, indexed-searcher, migrator, node-exporter, opentelemetry-collector, postgres-12-alpine, postgres_exporter, precise-code-intel-worker, prometheus, prometheus-gcp, redis-cache, redis-store, redis_exporter, repo-updater, search-indexer, searcher, worker
- Upload build trace
=======
- **End-to-end tests**: Executors E2E, Sourcegraph E2E, Sourcegraph QA, Sourcegraph Cluster (deploy-sourcegraph) QA, Sourcegraph Upgrade
- **Publish images**: alpine-3.14, cadvisor, codeinsights-db, codeintel-db, frontend, github-proxy, gitserver, grafana, indexed-searcher, jaeger-agent, jaeger-all-in-one, blobstore, blobstore2, node-exporter, postgres-12-alpine, postgres_exporter, precise-code-intel-worker, prometheus, prometheus-gcp, redis-cache, redis-store, redis_exporter, repo-updater, search-indexer, searcher, symbols, syntax-highlighter, worker, migrator, executor, executor-kubernetes, executor-vm, batcheshelper, opentelemetry-collector, embeddings, dind, bundled-executor, server, sg, llm-proxy
>>>>>>> main

### Browser extension release build

The run type for branches matching `bext/release` (exact match).

Base pipeline (more steps might be included based on branch changes):

- ESLint (all)
- ESLint (web)
- Stylelint (all)
- Test (client/browser)
- Puppeteer tests for chrome extension
- Test (all)
- E2E for chrome extension
- Extension release
- Extension release
- npm Release

### VS Code extension release build

The run type for branches matching `vsce/release` (exact match).

Base pipeline (more steps might be included based on branch changes):

- ESLint (all)
- ESLint (web)
- Stylelint (all)
- Tests for VS Code extension
- Extension release

### Cody VS Code extension release build

The run type for branches matching `cody/release` (exact match).

Base pipeline (more steps might be included based on branch changes):

- ESLint (all)
- ESLint (web)
- Stylelint (all)
- Integration tests for the Cody VS Code extension
- Cody release

### Main branch

The run type for branches matching `main` (exact match).

Base pipeline (more steps might be included based on branch changes):

- **Metadata**: Pipeline metadata
- **Pipeline setup**: Trigger async
- **Image builds**: Build symbols, Build server, Build batcheshelper, Build blobstore2, Build bundled-executor, Build dind, Build embeddings, Build executor, Build executor-kubernetes, Build executor-vm, Build jaeger-agent, Build jaeger-all-in-one, Build llm-proxy, Build sg, Build alpine-3.14, Build blobstore, Build cadvisor, Build codeinsights-db, Build codeintel-db, Build frontend, Build github-proxy, Build gitserver, Build grafana, Build indexed-searcher, Build migrator, Build node-exporter, Build opentelemetry-collector, Build postgres-12-alpine, Build postgres_exporter, Build precise-code-intel-worker, Build prometheus, Build prometheus-gcp, Build redis-cache, Build redis-store, Build redis_exporter, Build repo-updater, Build search-indexer, Build searcher, Build worker
- **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
- **Linters and static analysis**: Run sg lint
- **Client checks**: Upload Storybook to Chromatic, Enterprise build, Build (client/jetbrains), Tests for VS Code extension, Integration tests for the Cody VS Code extension, ESLint (all), ESLint (web), Stylelint (all)
- **CI script tests**: test-trace-command.sh
- **Integration tests**: Backend integration tests (gRPC), Backend integration tests, Code Intel QA
<<<<<<< HEAD
- **End-to-end tests**: Executors E2E, Sourcegraph E2E, Sourcegraph Upgrade
- **Publish images**: batcheshelper, blobstore2, bundled-executor, dind, embeddings, executor, executor-kubernetes, executor-vm, jaeger-agent, jaeger-all-in-one, llm-proxy, sg, symbols, server, alpine-3.14, blobstore, cadvisor, codeinsights-db, codeintel-db, frontend, github-proxy, gitserver, grafana, indexed-searcher, migrator, node-exporter, opentelemetry-collector, postgres-12-alpine, postgres_exporter, precise-code-intel-worker, prometheus, prometheus-gcp, redis-cache, redis-store, redis_exporter, repo-updater, search-indexer, searcher, worker, Publish executor image, Publish executor binary
- Upload build trace
=======
- **End-to-end tests**: Executors E2E, Sourcegraph E2E, Sourcegraph QA, Sourcegraph Cluster (deploy-sourcegraph) QA, Sourcegraph Upgrade
- **Publish images**: alpine-3.14, cadvisor, codeinsights-db, codeintel-db, frontend, github-proxy, gitserver, grafana, indexed-searcher, jaeger-agent, jaeger-all-in-one, blobstore, blobstore2, node-exporter, postgres-12-alpine, postgres_exporter, precise-code-intel-worker, prometheus, prometheus-gcp, redis-cache, redis-store, redis_exporter, repo-updater, search-indexer, searcher, symbols, syntax-highlighter, worker, migrator, executor, executor-kubernetes, executor-vm, batcheshelper, opentelemetry-collector, embeddings, dind, bundled-executor, server, sg, llm-proxy, Publish executor image, Publish executor binary
>>>>>>> main

### Main dry run

The run type for branches matching `main-dry-run/`.
You can create a build of this run type for your changes using:

```sh
sg ci build main-dry-run
```

Base pipeline (more steps might be included based on branch changes):

- **Metadata**: Pipeline metadata
- **Pipeline setup**: Trigger async
- **Image builds**: Build symbols, Build server, Build batcheshelper, Build blobstore2, Build bundled-executor, Build dind, Build embeddings, Build executor, Build executor-kubernetes, Build executor-vm, Build jaeger-agent, Build jaeger-all-in-one, Build llm-proxy, Build sg, Build alpine-3.14, Build blobstore, Build cadvisor, Build codeinsights-db, Build codeintel-db, Build frontend, Build github-proxy, Build gitserver, Build grafana, Build indexed-searcher, Build migrator, Build node-exporter, Build opentelemetry-collector, Build postgres-12-alpine, Build postgres_exporter, Build precise-code-intel-worker, Build prometheus, Build prometheus-gcp, Build redis-cache, Build redis-store, Build redis_exporter, Build repo-updater, Build search-indexer, Build searcher, Build worker
- **Bazel**: Ensure buildfiles are up to date, Tests, BackCompat Tests
- **Linters and static analysis**: Run sg lint
- **Client checks**: Upload Storybook to Chromatic, Enterprise build, Build (client/jetbrains), Tests for VS Code extension, Integration tests for the Cody VS Code extension, ESLint (all), ESLint (web), Stylelint (all)
- **CI script tests**: test-trace-command.sh
- **Integration tests**: Backend integration tests (gRPC), Backend integration tests, Code Intel QA
<<<<<<< HEAD
- **End-to-end tests**: Executors E2E, Sourcegraph E2E, Sourcegraph Upgrade
- **Publish images**: batcheshelper, blobstore2, bundled-executor, dind, embeddings, executor, executor-kubernetes, executor-vm, jaeger-agent, jaeger-all-in-one, llm-proxy, sg, symbols, server, alpine-3.14, blobstore, cadvisor, codeinsights-db, codeintel-db, frontend, github-proxy, gitserver, grafana, indexed-searcher, migrator, node-exporter, opentelemetry-collector, postgres-12-alpine, postgres_exporter, precise-code-intel-worker, prometheus, prometheus-gcp, redis-cache, redis-store, redis_exporter, repo-updater, search-indexer, searcher, worker
- Upload build trace
=======
- **End-to-end tests**: Executors E2E, Sourcegraph E2E, Sourcegraph QA, Sourcegraph Cluster (deploy-sourcegraph) QA, Sourcegraph Upgrade
- **Publish images**: alpine-3.14, cadvisor, codeinsights-db, codeintel-db, frontend, github-proxy, gitserver, grafana, indexed-searcher, jaeger-agent, jaeger-all-in-one, blobstore, blobstore2, node-exporter, postgres-12-alpine, postgres_exporter, precise-code-intel-worker, prometheus, prometheus-gcp, redis-cache, redis-store, redis_exporter, repo-updater, search-indexer, searcher, symbols, syntax-highlighter, worker, migrator, executor, executor-kubernetes, executor-vm, batcheshelper, opentelemetry-collector, embeddings, dind, bundled-executor, server, sg, llm-proxy
>>>>>>> main

### Patch image

The run type for branches matching `docker-images-patch/`, requires a branch argument in the second branch path segment.
You can create a build of this run type for your changes using:

```sh
sg ci build docker-images-patch
```

### Patch image without testing

The run type for branches matching `docker-images-patch-notest/`, requires a branch argument in the second branch path segment.
You can create a build of this run type for your changes using:

```sh
sg ci build docker-images-patch-notest
```

### Build all candidates without testing

The run type for branches matching `docker-images-candidates-notest/`.
You can create a build of this run type for your changes using:

```sh
sg ci build docker-images-candidates-notest
```

Base pipeline (more steps might be included based on branch changes):

- **Metadata**: Pipeline metadata
<<<<<<< HEAD
- **Image builds**: Build Docker images
- **Publish images**: Publish images
- Upload build trace
=======
- **Image builds**: Build alpine-3.14, Build cadvisor, Build codeinsights-db, Build codeintel-db, Build frontend, Build github-proxy, Build gitserver, Build grafana, Build indexed-searcher, Build jaeger-agent, Build jaeger-all-in-one, Build blobstore, Build blobstore2, Build node-exporter, Build postgres-12-alpine, Build postgres_exporter, Build precise-code-intel-worker, Build prometheus, Build prometheus-gcp, Build redis-cache, Build redis-store, Build redis_exporter, Build repo-updater, Build search-indexer, Build searcher, Build symbols, Build syntax-highlighter, Build worker, Build migrator, Build executor, Build executor-kubernetes, Build executor-vm, Build batcheshelper, Build opentelemetry-collector, Build embeddings, Build dind, Build bundled-executor, Build server, Build sg, Build llm-proxy
- **Publish images**: alpine-3.14, cadvisor, codeinsights-db, codeintel-db, frontend, github-proxy, gitserver, grafana, indexed-searcher, jaeger-agent, jaeger-all-in-one, blobstore, blobstore2, node-exporter, postgres-12-alpine, postgres_exporter, precise-code-intel-worker, prometheus, prometheus-gcp, redis-cache, redis-store, redis_exporter, repo-updater, search-indexer, searcher, symbols, syntax-highlighter, worker, migrator, executor, executor-kubernetes, executor-vm, batcheshelper, opentelemetry-collector, embeddings, dind, bundled-executor, server, sg, llm-proxy
>>>>>>> main

### Build executor without testing

The run type for branches matching `executor-patch-notest/`.
You can create a build of this run type for your changes using:

```sh
sg ci build executor-patch-notest
```

Base pipeline (more steps might be included based on branch changes):

- Build Docker images
- Scan executor-vm
- Build executor image
- Build docker registry mirror image
- Build executor binary
- Publish images
- Publish executor image
- Publish docker registry mirror image
- Publish executor binary

### Backend integration tests

The run type for branches matching `backend-integration/`.
You can create a build of this run type for your changes using:

```sh
sg ci build backend-integration
```

Base pipeline (more steps might be included based on branch changes):

- **Metadata**: Pipeline metadata
- Build Docker images
- Backend integration tests (gRPC)
- Backend integration tests
- **Linters and static analysis**: Run sg lint
- **Go checks**: Test (all), Test (all (gRPC)), Test (enterprise/internal/insights), Test (enterprise/internal/insights (gRPC)), Test (internal/repos), Test (internal/repos (gRPC)), Test (enterprise/internal/batches), Test (enterprise/internal/batches (gRPC)), Test (cmd/frontend), Test (cmd/frontend (gRPC)), Test (enterprise/cmd/frontend/internal/batches/resolvers), Test (enterprise/cmd/frontend/internal/batches/resolvers (gRPC)), Test (dev/sg), Test (dev/sg (gRPC)), Test (internal/database), Test (enterprise/internal/database), Build
