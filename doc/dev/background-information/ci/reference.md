<!-- DO NOT EDIT: generated via: go generate ./enterprise/dev/ci -->

# Pipeline types reference

This is a reference outlining what CI pipelines we generate under different conditions.

To preview the pipeline for your branch, use `sg ci preview`.

For a higher-level overview, please refer to the [continuous integration docs](https://docs.sourcegraph.com/dev/background-information/continuous_integration).

## Run types

### Pull request

The default run type.

- Pipeline for `Go` changes:
  - **Linters and static analysis**: Prettier, Misc Linters
  - **Go checks**: Test (all), Test (enterprise/internal/codeintel/stores/dbstore), Test (enterprise/internal/codeintel/stores/lsifstore), Test (enterprise/internal/insights), Test (internal/database), Test (internal/repos), Test (enterprise/internal/batches), Test (cmd/frontend), Test (enterprise/internal/database), Test (enterprise/cmd/frontend/internal/batches/resolvers), Build
  - Upload build trace

- Pipeline for `Client` changes:
  - **Pipeline setup**: Trigger async
  - **Linters and static analysis**: Prettier, Misc Linters, Yarn deduplicate lint
  - **Client checks**: Puppeteer tests prep, Puppeteer tests chunk #1, Puppeteer tests chunk #2, Puppeteer tests chunk #3, Puppeteer tests chunk #4, Puppeteer tests chunk #5, Puppeteer tests chunk #6, Puppeteer tests chunk #7, Puppeteer tests chunk #8, Puppeteer tests chunk #9, Puppeteer tests finalize, Upload Storybook to Chromatic, Test shared client code, Test wildcard client code, Build, Enterprise build, Test, Puppeteer tests for chrome extension, Test browser extension, Test branded client code, Typescript eslint, Stylelint
  - Upload build trace

- Pipeline for `GraphQL` changes:
  - **Linters and static analysis**: Prettier, Misc Linters, GraphQL lint
  - **Client checks**: Puppeteer tests prep, Puppeteer tests chunk #1, Puppeteer tests chunk #2, Puppeteer tests chunk #3, Puppeteer tests chunk #4, Puppeteer tests chunk #5, Puppeteer tests chunk #6, Puppeteer tests chunk #7, Puppeteer tests chunk #8, Puppeteer tests chunk #9, Puppeteer tests finalize, Upload Storybook to Chromatic, Test shared client code, Test wildcard client code, Build, Enterprise build, Test, Puppeteer tests for chrome extension, Test browser extension, Test branded client code, Typescript eslint, Stylelint
  - **Go checks**: Test (all), Test (enterprise/internal/codeintel/stores/dbstore), Test (enterprise/internal/codeintel/stores/lsifstore), Test (enterprise/internal/insights), Test (internal/database), Test (internal/repos), Test (enterprise/internal/batches), Test (cmd/frontend), Test (enterprise/internal/database), Test (enterprise/cmd/frontend/internal/batches/resolvers), Build
  - Upload build trace

- Pipeline for `DatabaseSchema` changes:
  - **Linters and static analysis**: Prettier, Misc Linters
  - **DB backcompat tests**: Backcompat test (all), Backcompat test (enterprise/internal/codeintel/stores/dbstore), Backcompat test (enterprise/internal/codeintel/stores/lsifstore), Backcompat test (enterprise/internal/insights), Backcompat test (internal/database), Backcompat test (internal/repos), Backcompat test (enterprise/internal/batches), Backcompat test (cmd/frontend), Backcompat test (enterprise/internal/database), Backcompat test (enterprise/cmd/frontend/internal/batches/resolvers)
  - Upload build trace

- Pipeline for `Docs` changes:
  - **Linters and static analysis**: Prettier, Misc Linters, Check and build docsite
  - Upload build trace

- Pipeline for `Dockerfiles` changes:
  - **Linters and static analysis**: Prettier, Misc Linters, Lint
  - Upload build trace

- Pipeline for `ExecutorDockerRegistryMirror` changes:
  - **Linters and static analysis**: Prettier, Misc Linters
  - Upload build trace

- Pipeline for `CIScripts` changes:
  - **Linters and static analysis**: Prettier, Misc Linters
  - **CI script tests**: test-trace-command.sh
  - Upload build trace

- Pipeline for `Terraform` changes:
  - **Linters and static analysis**: Prettier, Misc Linters, security - checkov
  - Upload build trace

- Pipeline for `SVG` changes:
  - **Linters and static analysis**: Prettier, Misc Linters, SVG lint
  - Upload build trace

### Release branch nightly healthcheck build

The run type for environment including `{"RELEASE_NIGHTLY":"true"}`.

- Default pipeline:
  - **Pipeline setup**: Trigger async
  - **Linters and static analysis**: Prettier, Misc Linters, GraphQL lint, SVG lint, Yarn deduplicate lint, Lint, security - checkov, Check and build docsite
  - **Client checks**: Puppeteer tests prep, Puppeteer tests chunk #1, Puppeteer tests chunk #2, Puppeteer tests chunk #3, Puppeteer tests chunk #4, Puppeteer tests chunk #5, Puppeteer tests chunk #6, Puppeteer tests chunk #7, Puppeteer tests chunk #8, Puppeteer tests chunk #9, Puppeteer tests finalize, Upload Storybook to Chromatic, Test shared client code, Test wildcard client code, Build, Enterprise build, Test, Puppeteer tests for chrome extension, Test browser extension, Test branded client code, Typescript eslint, Stylelint
  - **Go checks**: Test (all), Test (enterprise/internal/codeintel/stores/dbstore), Test (enterprise/internal/codeintel/stores/lsifstore), Test (enterprise/internal/insights), Test (internal/database), Test (internal/repos), Test (enterprise/internal/batches), Test (cmd/frontend), Test (enterprise/internal/database), Test (enterprise/cmd/frontend/internal/batches/resolvers), Build
  - **DB backcompat tests**: Backcompat test (all), Backcompat test (enterprise/internal/codeintel/stores/dbstore), Backcompat test (enterprise/internal/codeintel/stores/lsifstore), Backcompat test (enterprise/internal/insights), Backcompat test (internal/database), Backcompat test (internal/repos), Backcompat test (enterprise/internal/batches), Backcompat test (cmd/frontend), Backcompat test (enterprise/internal/database), Backcompat test (enterprise/cmd/frontend/internal/batches/resolvers)
  - **CI script tests**: test-trace-command.sh
  - Upload build trace

### Browser extension nightly release build

The run type for environment including `{"BEXT_NIGHTLY":"true"}`.

- Default pipeline:
  - **Pipeline setup**: Trigger async
  - **Linters and static analysis**: Prettier, Misc Linters, GraphQL lint, SVG lint, Yarn deduplicate lint, Lint, security - checkov, Check and build docsite
  - **Client checks**: Puppeteer tests prep, Puppeteer tests chunk #1, Puppeteer tests chunk #2, Puppeteer tests chunk #3, Puppeteer tests chunk #4, Puppeteer tests chunk #5, Puppeteer tests chunk #6, Puppeteer tests chunk #7, Puppeteer tests chunk #8, Puppeteer tests chunk #9, Puppeteer tests finalize, Upload Storybook to Chromatic, Test shared client code, Test wildcard client code, Build, Enterprise build, Test, Puppeteer tests for chrome extension, Test browser extension, Test branded client code, Typescript eslint, Stylelint
  - **Go checks**: Test (all), Test (enterprise/internal/codeintel/stores/dbstore), Test (enterprise/internal/codeintel/stores/lsifstore), Test (enterprise/internal/insights), Test (internal/database), Test (internal/repos), Test (enterprise/internal/batches), Test (cmd/frontend), Test (enterprise/internal/database), Test (enterprise/cmd/frontend/internal/batches/resolvers), Build
  - **DB backcompat tests**: Backcompat test (all), Backcompat test (enterprise/internal/codeintel/stores/dbstore), Backcompat test (enterprise/internal/codeintel/stores/lsifstore), Backcompat test (enterprise/internal/insights), Backcompat test (internal/database), Backcompat test (internal/repos), Backcompat test (enterprise/internal/batches), Backcompat test (cmd/frontend), Backcompat test (enterprise/internal/database), Backcompat test (enterprise/cmd/frontend/internal/batches/resolvers)
  - **CI script tests**: test-trace-command.sh
  - Upload build trace

### Tagged release

The run type for tags starting with `v`.

- Default pipeline:
  - **Pipeline setup**: Trigger async
  - **Linters and static analysis**: Prettier, Misc Linters, GraphQL lint, SVG lint, Yarn deduplicate lint, Lint, security - checkov, Check and build docsite
  - **Client checks**: Puppeteer tests prep, Puppeteer tests chunk #1, Puppeteer tests chunk #2, Puppeteer tests chunk #3, Puppeteer tests chunk #4, Puppeteer tests chunk #5, Puppeteer tests chunk #6, Puppeteer tests chunk #7, Puppeteer tests chunk #8, Puppeteer tests chunk #9, Puppeteer tests finalize, Upload Storybook to Chromatic, Test shared client code, Test wildcard client code, Build, Enterprise build, Test, Puppeteer tests for chrome extension, Test browser extension, Test branded client code, Typescript eslint, Stylelint
  - **Go checks**: Test (all), Test (enterprise/internal/codeintel/stores/dbstore), Test (enterprise/internal/codeintel/stores/lsifstore), Test (enterprise/internal/insights), Test (internal/database), Test (internal/repos), Test (enterprise/internal/batches), Test (cmd/frontend), Test (enterprise/internal/database), Test (enterprise/cmd/frontend/internal/batches/resolvers), Build
  - **DB backcompat tests**: Backcompat test (all), Backcompat test (enterprise/internal/codeintel/stores/dbstore), Backcompat test (enterprise/internal/codeintel/stores/lsifstore), Backcompat test (enterprise/internal/insights), Backcompat test (internal/database), Backcompat test (internal/repos), Backcompat test (enterprise/internal/batches), Backcompat test (cmd/frontend), Backcompat test (enterprise/internal/database), Backcompat test (enterprise/cmd/frontend/internal/batches/resolvers)
  - **CI script tests**: test-trace-command.sh
  - Upload build trace

### Release branch

The run type for branches matching `^[0-9]+\.[0-9]+$` (regexp).

- Default pipeline:
  - **Pipeline setup**: Trigger async
  - **Linters and static analysis**: Prettier, Misc Linters, GraphQL lint, SVG lint, Yarn deduplicate lint, Lint, security - checkov, Check and build docsite
  - **Client checks**: Puppeteer tests prep, Puppeteer tests chunk #1, Puppeteer tests chunk #2, Puppeteer tests chunk #3, Puppeteer tests chunk #4, Puppeteer tests chunk #5, Puppeteer tests chunk #6, Puppeteer tests chunk #7, Puppeteer tests chunk #8, Puppeteer tests chunk #9, Puppeteer tests finalize, Upload Storybook to Chromatic, Test shared client code, Test wildcard client code, Build, Enterprise build, Test, Puppeteer tests for chrome extension, Test browser extension, Test branded client code, Typescript eslint, Stylelint
  - **Go checks**: Test (all), Test (enterprise/internal/codeintel/stores/dbstore), Test (enterprise/internal/codeintel/stores/lsifstore), Test (enterprise/internal/insights), Test (internal/database), Test (internal/repos), Test (enterprise/internal/batches), Test (cmd/frontend), Test (enterprise/internal/database), Test (enterprise/cmd/frontend/internal/batches/resolvers), Build
  - **DB backcompat tests**: Backcompat test (all), Backcompat test (enterprise/internal/codeintel/stores/dbstore), Backcompat test (enterprise/internal/codeintel/stores/lsifstore), Backcompat test (enterprise/internal/insights), Backcompat test (internal/database), Backcompat test (internal/repos), Backcompat test (enterprise/internal/batches), Backcompat test (cmd/frontend), Backcompat test (enterprise/internal/database), Backcompat test (enterprise/cmd/frontend/internal/batches/resolvers)
  - **CI script tests**: test-trace-command.sh
  - Upload build trace

### Browser extension release build

The run type for branches matching `bext/release` (exact).

- Default pipeline:
  - **Pipeline setup**: Trigger async
  - **Linters and static analysis**: Prettier, Misc Linters, GraphQL lint, SVG lint, Yarn deduplicate lint, Lint, security - checkov, Check and build docsite
  - **Client checks**: Puppeteer tests prep, Puppeteer tests chunk #1, Puppeteer tests chunk #2, Puppeteer tests chunk #3, Puppeteer tests chunk #4, Puppeteer tests chunk #5, Puppeteer tests chunk #6, Puppeteer tests chunk #7, Puppeteer tests chunk #8, Puppeteer tests chunk #9, Puppeteer tests finalize, Upload Storybook to Chromatic, Test shared client code, Test wildcard client code, Build, Enterprise build, Test, Puppeteer tests for chrome extension, Test browser extension, Test branded client code, Typescript eslint, Stylelint
  - **Go checks**: Test (all), Test (enterprise/internal/codeintel/stores/dbstore), Test (enterprise/internal/codeintel/stores/lsifstore), Test (enterprise/internal/insights), Test (internal/database), Test (internal/repos), Test (enterprise/internal/batches), Test (cmd/frontend), Test (enterprise/internal/database), Test (enterprise/cmd/frontend/internal/batches/resolvers), Build
  - **DB backcompat tests**: Backcompat test (all), Backcompat test (enterprise/internal/codeintel/stores/dbstore), Backcompat test (enterprise/internal/codeintel/stores/lsifstore), Backcompat test (enterprise/internal/insights), Backcompat test (internal/database), Backcompat test (internal/repos), Backcompat test (enterprise/internal/batches), Backcompat test (cmd/frontend), Backcompat test (enterprise/internal/database), Backcompat test (enterprise/cmd/frontend/internal/batches/resolvers)
  - **CI script tests**: test-trace-command.sh
  - Upload build trace

### Main branch

The run type for branches matching `main` (exact).

- Default pipeline:
  - **Pipeline setup**: Trigger async
  - **Linters and static analysis**: Prettier, Misc Linters, GraphQL lint, SVG lint, Yarn deduplicate lint, Lint, security - checkov, Check and build docsite
  - **Client checks**: Puppeteer tests prep, Puppeteer tests chunk #1, Puppeteer tests chunk #2, Puppeteer tests chunk #3, Puppeteer tests chunk #4, Puppeteer tests chunk #5, Puppeteer tests chunk #6, Puppeteer tests chunk #7, Puppeteer tests chunk #8, Puppeteer tests chunk #9, Puppeteer tests finalize, Upload Storybook to Chromatic, Test shared client code, Test wildcard client code, Build, Enterprise build, Test, Puppeteer tests for chrome extension, Test browser extension, Test branded client code, Typescript eslint, Stylelint
  - **Go checks**: Test (all), Test (enterprise/internal/codeintel/stores/dbstore), Test (enterprise/internal/codeintel/stores/lsifstore), Test (enterprise/internal/insights), Test (internal/database), Test (internal/repos), Test (enterprise/internal/batches), Test (cmd/frontend), Test (enterprise/internal/database), Test (enterprise/cmd/frontend/internal/batches/resolvers), Build
  - **DB backcompat tests**: Backcompat test (all), Backcompat test (enterprise/internal/codeintel/stores/dbstore), Backcompat test (enterprise/internal/codeintel/stores/lsifstore), Backcompat test (enterprise/internal/insights), Backcompat test (internal/database), Backcompat test (internal/repos), Backcompat test (enterprise/internal/batches), Backcompat test (cmd/frontend), Backcompat test (enterprise/internal/database), Backcompat test (enterprise/cmd/frontend/internal/batches/resolvers)
  - **CI script tests**: test-trace-command.sh
  - Upload build trace

### Main dry run

The run type for branches matching `main-dry-run/`.

- Default pipeline:
  - **Pipeline setup**: Trigger async
  - **Linters and static analysis**: Prettier, Misc Linters, GraphQL lint, SVG lint, Yarn deduplicate lint, Lint, security - checkov, Check and build docsite
  - **Client checks**: Puppeteer tests prep, Puppeteer tests chunk #1, Puppeteer tests chunk #2, Puppeteer tests chunk #3, Puppeteer tests chunk #4, Puppeteer tests chunk #5, Puppeteer tests chunk #6, Puppeteer tests chunk #7, Puppeteer tests chunk #8, Puppeteer tests chunk #9, Puppeteer tests finalize, Upload Storybook to Chromatic, Test shared client code, Test wildcard client code, Build, Enterprise build, Test, Puppeteer tests for chrome extension, Test browser extension, Test branded client code, Typescript eslint, Stylelint
  - **Go checks**: Test (all), Test (enterprise/internal/codeintel/stores/dbstore), Test (enterprise/internal/codeintel/stores/lsifstore), Test (enterprise/internal/insights), Test (internal/database), Test (internal/repos), Test (enterprise/internal/batches), Test (cmd/frontend), Test (enterprise/internal/database), Test (enterprise/cmd/frontend/internal/batches/resolvers), Build
  - **DB backcompat tests**: Backcompat test (all), Backcompat test (enterprise/internal/codeintel/stores/dbstore), Backcompat test (enterprise/internal/codeintel/stores/lsifstore), Backcompat test (enterprise/internal/insights), Backcompat test (internal/database), Backcompat test (internal/repos), Backcompat test (enterprise/internal/batches), Backcompat test (cmd/frontend), Backcompat test (enterprise/internal/database), Backcompat test (enterprise/cmd/frontend/internal/batches/resolvers)
  - **CI script tests**: test-trace-command.sh
  - Upload build trace

### Patch image

The run type for branches matching `docker-images-patch/`.

- Default pipeline:
  - **Pipeline setup**: Trigger async
  - **Linters and static analysis**: Prettier, Misc Linters, GraphQL lint, SVG lint, Yarn deduplicate lint, Lint, security - checkov, Check and build docsite
  - **Client checks**: Puppeteer tests prep, Puppeteer tests chunk #1, Puppeteer tests chunk #2, Puppeteer tests chunk #3, Puppeteer tests chunk #4, Puppeteer tests chunk #5, Puppeteer tests chunk #6, Puppeteer tests chunk #7, Puppeteer tests chunk #8, Puppeteer tests chunk #9, Puppeteer tests finalize, Upload Storybook to Chromatic, Test shared client code, Test wildcard client code, Build, Enterprise build, Test, Puppeteer tests for chrome extension, Test browser extension, Test branded client code, Typescript eslint, Stylelint
  - **Go checks**: Test (all), Test (enterprise/internal/codeintel/stores/dbstore), Test (enterprise/internal/codeintel/stores/lsifstore), Test (enterprise/internal/insights), Test (internal/database), Test (internal/repos), Test (enterprise/internal/batches), Test (cmd/frontend), Test (enterprise/internal/database), Test (enterprise/cmd/frontend/internal/batches/resolvers), Build
  - **DB backcompat tests**: Backcompat test (all), Backcompat test (enterprise/internal/codeintel/stores/dbstore), Backcompat test (enterprise/internal/codeintel/stores/lsifstore), Backcompat test (enterprise/internal/insights), Backcompat test (internal/database), Backcompat test (internal/repos), Backcompat test (enterprise/internal/batches), Backcompat test (cmd/frontend), Backcompat test (enterprise/internal/database), Backcompat test (enterprise/cmd/frontend/internal/batches/resolvers)
  - **CI script tests**: test-trace-command.sh
  - Upload build trace

### Patch image without testing

The run type for branches matching `docker-images-patch-notest/`.

- Default pipeline:
  - **Pipeline setup**: Trigger async
  - **Linters and static analysis**: Prettier, Misc Linters, GraphQL lint, SVG lint, Yarn deduplicate lint, Lint, security - checkov, Check and build docsite
  - **Client checks**: Puppeteer tests prep, Puppeteer tests chunk #1, Puppeteer tests chunk #2, Puppeteer tests chunk #3, Puppeteer tests chunk #4, Puppeteer tests chunk #5, Puppeteer tests chunk #6, Puppeteer tests chunk #7, Puppeteer tests chunk #8, Puppeteer tests chunk #9, Puppeteer tests finalize, Upload Storybook to Chromatic, Test shared client code, Test wildcard client code, Build, Enterprise build, Test, Puppeteer tests for chrome extension, Test browser extension, Test branded client code, Typescript eslint, Stylelint
  - **Go checks**: Test (all), Test (enterprise/internal/codeintel/stores/dbstore), Test (enterprise/internal/codeintel/stores/lsifstore), Test (enterprise/internal/insights), Test (internal/database), Test (internal/repos), Test (enterprise/internal/batches), Test (cmd/frontend), Test (enterprise/internal/database), Test (enterprise/cmd/frontend/internal/batches/resolvers), Build
  - **DB backcompat tests**: Backcompat test (all), Backcompat test (enterprise/internal/codeintel/stores/dbstore), Backcompat test (enterprise/internal/codeintel/stores/lsifstore), Backcompat test (enterprise/internal/insights), Backcompat test (internal/database), Backcompat test (internal/repos), Backcompat test (enterprise/internal/batches), Backcompat test (cmd/frontend), Backcompat test (enterprise/internal/database), Backcompat test (enterprise/cmd/frontend/internal/batches/resolvers)
  - **CI script tests**: test-trace-command.sh
  - Upload build trace

### Build all candidates without testing

The run type for branches matching `docker-images-candidates-notest/`.

- Default pipeline:
  - **Pipeline setup**: Trigger async
  - **Linters and static analysis**: Prettier, Misc Linters, GraphQL lint, SVG lint, Yarn deduplicate lint, Lint, security - checkov, Check and build docsite
  - **Client checks**: Puppeteer tests prep, Puppeteer tests chunk #1, Puppeteer tests chunk #2, Puppeteer tests chunk #3, Puppeteer tests chunk #4, Puppeteer tests chunk #5, Puppeteer tests chunk #6, Puppeteer tests chunk #7, Puppeteer tests chunk #8, Puppeteer tests chunk #9, Puppeteer tests finalize, Upload Storybook to Chromatic, Test shared client code, Test wildcard client code, Build, Enterprise build, Test, Puppeteer tests for chrome extension, Test browser extension, Test branded client code, Typescript eslint, Stylelint
  - **Go checks**: Test (all), Test (enterprise/internal/codeintel/stores/dbstore), Test (enterprise/internal/codeintel/stores/lsifstore), Test (enterprise/internal/insights), Test (internal/database), Test (internal/repos), Test (enterprise/internal/batches), Test (cmd/frontend), Test (enterprise/internal/database), Test (enterprise/cmd/frontend/internal/batches/resolvers), Build
  - **DB backcompat tests**: Backcompat test (all), Backcompat test (enterprise/internal/codeintel/stores/dbstore), Backcompat test (enterprise/internal/codeintel/stores/lsifstore), Backcompat test (enterprise/internal/insights), Backcompat test (internal/database), Backcompat test (internal/repos), Backcompat test (enterprise/internal/batches), Backcompat test (cmd/frontend), Backcompat test (enterprise/internal/database), Backcompat test (enterprise/cmd/frontend/internal/batches/resolvers)
  - **CI script tests**: test-trace-command.sh
  - Upload build trace

### Build executor without testing

The run type for branches matching `executor-patch-notest/`.

- Default pipeline:
  - **Pipeline setup**: Trigger async
  - **Linters and static analysis**: Prettier, Misc Linters, GraphQL lint, SVG lint, Yarn deduplicate lint, Lint, security - checkov, Check and build docsite
  - **Client checks**: Puppeteer tests prep, Puppeteer tests chunk #1, Puppeteer tests chunk #2, Puppeteer tests chunk #3, Puppeteer tests chunk #4, Puppeteer tests chunk #5, Puppeteer tests chunk #6, Puppeteer tests chunk #7, Puppeteer tests chunk #8, Puppeteer tests chunk #9, Puppeteer tests finalize, Upload Storybook to Chromatic, Test shared client code, Test wildcard client code, Build, Enterprise build, Test, Puppeteer tests for chrome extension, Test browser extension, Test branded client code, Typescript eslint, Stylelint
  - **Go checks**: Test (all), Test (enterprise/internal/codeintel/stores/dbstore), Test (enterprise/internal/codeintel/stores/lsifstore), Test (enterprise/internal/insights), Test (internal/database), Test (internal/repos), Test (enterprise/internal/batches), Test (cmd/frontend), Test (enterprise/internal/database), Test (enterprise/cmd/frontend/internal/batches/resolvers), Build
  - **DB backcompat tests**: Backcompat test (all), Backcompat test (enterprise/internal/codeintel/stores/dbstore), Backcompat test (enterprise/internal/codeintel/stores/lsifstore), Backcompat test (enterprise/internal/insights), Backcompat test (internal/database), Backcompat test (internal/repos), Backcompat test (enterprise/internal/batches), Backcompat test (cmd/frontend), Backcompat test (enterprise/internal/database), Backcompat test (enterprise/cmd/frontend/internal/batches/resolvers)
  - **CI script tests**: test-trace-command.sh
  - Upload build trace

### Backend integration tests

The run type for branches matching `backend-integration/`.

- Default pipeline:
  - **Pipeline setup**: Trigger async
  - **Linters and static analysis**: Prettier, Misc Linters, GraphQL lint, SVG lint, Yarn deduplicate lint, Lint, security - checkov, Check and build docsite
  - **Client checks**: Puppeteer tests prep, Puppeteer tests chunk #1, Puppeteer tests chunk #2, Puppeteer tests chunk #3, Puppeteer tests chunk #4, Puppeteer tests chunk #5, Puppeteer tests chunk #6, Puppeteer tests chunk #7, Puppeteer tests chunk #8, Puppeteer tests chunk #9, Puppeteer tests finalize, Upload Storybook to Chromatic, Test shared client code, Test wildcard client code, Build, Enterprise build, Test, Puppeteer tests for chrome extension, Test browser extension, Test branded client code, Typescript eslint, Stylelint
  - **Go checks**: Test (all), Test (enterprise/internal/codeintel/stores/dbstore), Test (enterprise/internal/codeintel/stores/lsifstore), Test (enterprise/internal/insights), Test (internal/database), Test (internal/repos), Test (enterprise/internal/batches), Test (cmd/frontend), Test (enterprise/internal/database), Test (enterprise/cmd/frontend/internal/batches/resolvers), Build
  - **DB backcompat tests**: Backcompat test (all), Backcompat test (enterprise/internal/codeintel/stores/dbstore), Backcompat test (enterprise/internal/codeintel/stores/lsifstore), Backcompat test (enterprise/internal/insights), Backcompat test (internal/database), Backcompat test (internal/repos), Backcompat test (enterprise/internal/batches), Backcompat test (cmd/frontend), Backcompat test (enterprise/internal/database), Backcompat test (enterprise/cmd/frontend/internal/batches/resolvers)
  - **CI script tests**: test-trace-command.sh
  - Upload build trace
