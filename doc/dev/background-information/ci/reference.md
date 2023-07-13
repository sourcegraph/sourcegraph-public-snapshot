<!-- DO NOT EDIT: generated via: go generate ./enterprise/dev/ci -->

# Pipeline types reference

This is a reference outlining what CI pipelines we generate under different conditions.

To preview the pipeline for your branch, use `sg ci preview`.

For a higher-level overview, please refer to the [continuous integration docs](https://docs.sourcegraph.com/dev/background-information/ci).

## Run types

### Pull request

The default run type.

- Pipeline for `Go` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests
  - **Linters and static analysis**: Run sg lint

- Pipeline for `Client` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests
  - **Linters and static analysis**: Run sg lint
  - **Client checks**: Upload Storybook to Chromatic, Enterprise build, Build (client/jetbrains), Tests for VS Code extension, Unit and integration tests for the Cody VS Code extension, E2E tests for the Cody VS Code extension, Stylelint (all)
  - **Pipeline setup**: Trigger async

- Pipeline for `GraphQL` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests
  - **Client checks**: Upload Storybook to Chromatic, Enterprise build, Build (client/jetbrains), Tests for VS Code extension, Unit and integration tests for the Cody VS Code extension, E2E tests for the Cody VS Code extension, Stylelint (all)

- Pipeline for `DatabaseSchema` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests

- Pipeline for `Docs` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests
  - **Linters and static analysis**: Run sg lint

- Pipeline for `Dockerfiles` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests
  - **Linters and static analysis**: Run sg lint

- Pipeline for `ExecutorVMImage` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests

- Pipeline for `ExecutorDockerRegistryMirror` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests

- Pipeline for `CIScripts` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests

- Pipeline for `Terraform` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests

- Pipeline for `SVG` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests
  - **Linters and static analysis**: Run sg lint

- Pipeline for `Shell` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests
  - **Linters and static analysis**: Run sg lint

- Pipeline for `DockerImages` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests

- Pipeline for `WolfiPackages` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests

- Pipeline for `WolfiBaseImages` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests

- Pipeline for `Protobuf` changes:
  - Ensure buildfiles are up to date
  - Tests
  - BackCompat Tests
  - **Linters and static analysis**: Run sg lint

### Wolfi Exp Branch

The run type for branches matching `wolfi/`.
You can create a build of this run type for your changes using:

```sh
sg ci build wolfi
```

Base pipeline (more steps might be included based on branch changes):


### Release branch nightly healthcheck build

The run type for environment including `{"RELEASE_NIGHTLY":"true"}`.

Base pipeline (more steps might be included based on branch changes):

- Trigger 5.1 release branch healthcheck build
- Trigger 5.0 release branch healthcheck build

### Browser extension nightly release build

The run type for environment including `{"BEXT_NIGHTLY":"true"}`.

Base pipeline (more steps might be included based on branch changes):

- Test (client/browser)
- Puppeteer tests for chrome extension
- Test (all)
- E2E for chrome extension

### VS Code extension nightly release build

The run type for environment including `{"VSCE_NIGHTLY":"true"}`.

Base pipeline (more steps might be included based on branch changes):

- Tests for VS Code extension

### Cody VS Code extension nightly release build

The run type for environment including `{"CODY_NIGHTLY":"true"}`.

Base pipeline (more steps might be included based on branch changes):

- Unit and integration tests for the Cody VS Code extension
- E2E tests for the Cody VS Code extension
- Cody release

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

- **Pipeline setup**: Trigger async
- **Image builds**: Build Docker images, Build executor image, Build executor binary, Build docker registry mirror image
- Ensure buildfiles are up to date
- Tests
- BackCompat Tests
- **Linters and static analysis**: Run sg lint
- **Client checks**: Upload Storybook to Chromatic, Enterprise build, Build (client/jetbrains), Tests for VS Code extension, Unit and integration tests for the Cody VS Code extension, E2E tests for the Cody VS Code extension, Stylelint (all)
- **Publish candidate images**: Push candidate Images
- **End-to-end tests**: Executors E2E
- **Publish images**: executor-vm, alpine-3.14, codeinsights-db, codeintel-db, postgres-12-alpine, prometheus-gcp, Publish executor image, Publish executor binary, Publish docker registry mirror image, Push final images

### Release branch

The run type for branches matching `^[0-9]+\.[0-9]+$` (regexp match).

Base pipeline (more steps might be included based on branch changes):

- **Pipeline setup**: Trigger async
- **Image builds**: Build Docker images, Build executor image, Build executor binary, Build docker registry mirror image
- Ensure buildfiles are up to date
- Tests
- BackCompat Tests
- **Linters and static analysis**: Run sg lint
- **Client checks**: Upload Storybook to Chromatic, Enterprise build, Build (client/jetbrains), Tests for VS Code extension, Unit and integration tests for the Cody VS Code extension, E2E tests for the Cody VS Code extension, Stylelint (all)
- **Publish candidate images**: Push candidate Images
- **End-to-end tests**: Executors E2E
- **Publish images**: executor-vm, alpine-3.14, codeinsights-db, codeintel-db, postgres-12-alpine, prometheus-gcp, Push final images

### Browser extension release build

The run type for branches matching `bext/release` (exact match).

Base pipeline (more steps might be included based on branch changes):

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

- Tests for VS Code extension
- Extension release

### Cody VS Code extension release build

The run type for branches matching `cody/release` (exact match).

Base pipeline (more steps might be included based on branch changes):

- Unit and integration tests for the Cody VS Code extension
- E2E tests for the Cody VS Code extension
- Cody release

### Main branch

The run type for branches matching `main` (exact match).

Base pipeline (more steps might be included based on branch changes):

- **Pipeline setup**: Trigger async
- **Image builds**: Build Docker images, Build executor image, Build executor binary
- Ensure buildfiles are up to date
- Tests
- BackCompat Tests
- **Linters and static analysis**: Run sg lint
- **Client checks**: Upload Storybook to Chromatic, Enterprise build, Build (client/jetbrains), Tests for VS Code extension, Unit and integration tests for the Cody VS Code extension, E2E tests for the Cody VS Code extension, Stylelint (all)
- **Publish candidate images**: Push candidate Images
- **End-to-end tests**: Executors E2E
- **Publish images**: executor-vm, alpine-3.14, codeinsights-db, codeintel-db, postgres-12-alpine, prometheus-gcp, Publish executor image, Publish executor binary, Push final images

### Main dry run

The run type for branches matching `main-dry-run/`.
You can create a build of this run type for your changes using:

```sh
sg ci build main-dry-run
```

Base pipeline (more steps might be included based on branch changes):

- **Pipeline setup**: Trigger async
- **Image builds**: Build Docker images, Build executor image, Build executor binary
- Ensure buildfiles are up to date
- Tests
- BackCompat Tests
- **Linters and static analysis**: Run sg lint
- **Client checks**: Upload Storybook to Chromatic, Enterprise build, Build (client/jetbrains), Tests for VS Code extension, Unit and integration tests for the Cody VS Code extension, E2E tests for the Cody VS Code extension, Stylelint (all)
- **Publish candidate images**: Push candidate Images
- **End-to-end tests**: Executors E2E
- **Publish images**: executor-vm, alpine-3.14, codeinsights-db, codeintel-db, postgres-12-alpine, prometheus-gcp, Push final images

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

### Build executor without testing

The run type for branches matching `executor-patch-notest/`.
You can create a build of this run type for your changes using:

```sh
sg ci build executor-patch-notest
```

Base pipeline (more steps might be included based on branch changes):

- Build executor-vm
- Scan executor-vm
- Build executor image
- Build docker registry mirror image
- Build executor binary
- executor-vm
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

- **Pipeline setup**: Trigger async
- **Image builds**: Build Docker images
- Ensure buildfiles are up to date
- Tests
- BackCompat Tests
- **Linters and static analysis**: Run sg lint
- **Client checks**: Upload Storybook to Chromatic, Enterprise build, Build (client/jetbrains), Tests for VS Code extension, Unit and integration tests for the Cody VS Code extension, E2E tests for the Cody VS Code extension, Stylelint (all)
- **Publish candidate images**: Push candidate Images
- **End-to-end tests**: Executors E2E
- **Publish images**: executor-vm, alpine-3.14, codeinsights-db, codeintel-db, postgres-12-alpine, prometheus-gcp, Push final images

### Bazel command

The run type for branches matching `bazel-do/`.
You can create a build of this run type for your changes using:

```sh
sg ci build bazel-do
```
