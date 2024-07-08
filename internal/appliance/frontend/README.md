# Sourcegraph Appliance Maintenance UI

## Build

This will produce the distributable `dist` folder in `bazel-bin/maintenance/dist`:

    bazel build //maintenance:build

## Local run

This will run the service locally, starting a Vite developer environment:

    bazel run //maintenance:start
