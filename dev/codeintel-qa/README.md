# Precise code intel tester

This package provides integration and load testing utilities for precise code intel services.

## Prerequisites

Ensure that the following tools are available on your path:

- [`src`](https://github.com/sourcegraph/src-cli)
- [`lsif-go`](https://github.com/sourcegraph/lsif-go)
- [`gsutil`](https://cloud.google.com/storage/docs/gsutil_install) (and authenticated to the `sourcegraph-dev` project)

Set:

```sh
SOURCEGRAPH_BASE_URL=http://localhost:3080
SOURCEGRAPH_SUDO_TOKEN=<YOUR SOURCEGRAPH API ACCESS TOKEN>
```

## Testing

1. Ensure these repositories exist on your instance (in `Site Admin` -> `Manage repositories` -> `GitHub`):

```
  "repos": [
    "sourcegraph-testing/etcd",
    "sourcegraph-testing/tidb",
    "sourcegraph-testing/titan",
    "sourcegraph-testing/zap"
  ],
```

2. Download the test indexes by running the following command:

```
./scripts/download.sh
```

Alternatively, generate them by running the following command (this takes much longer):

```
./scripts/clone-and-index.sh
```

If there is previous upload or index state on the target instance, they can be cleared by running the following command:

```
go run ./cmd/clear
```

Upload the indexes to the target instance by running the following command:

```
go run ./cmd/upload
```

Then run test queries against the target instance by running the following command:

```
go run ./cmd/query
```

## Refreshing indexes

If there is a change to an indexer that needs to be tested, the indexes can be regenerated and uploaded to gcloud for future test runs.

Generate indexes by running the following command:

```
./scripts/clone-and-index.sh
```

Upload the generated indexes by running the following command:

```
./scripts/upload.sh
```

Or if you just want to test an indexer change locally, you can:

```sh
rm -rf testdata/indexes/
```

Then rerun the testing steps described above (starting at `clone-and-index.sh`)
