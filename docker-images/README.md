# Sourcegraph derivative Docker images

This directory contains Sourcegraph docker images which are derivatives of an existing Docker image, but with better defaults for our use cases. For example:

- `sourcegraph/alpine` handles setting up a `sourcegraph` user account, installing common packages.
- `sourcegraph/postgres` is `postgres` but with some Sourcegraph defaults.

If you are looking for our non-derivative Docker images, see e.g. `/cmd/.../Dockerfile` and `/enterprise/.../frontend/Dockerfile` instead.

### Building

These images are not yet built on CI. To build one, you must sign in to our Docker Hub and run `make <image name>` in this directory. For example:

```Makefile
make alpine
```

Before running the above command, you should have your changes reviewed and merged into `master`.

### Known issues

(1) Many of our derivative images have not yet been moved here from our [private infrastructure repository](https://github.com/sourcegraph/infrastructure/tree/master/docker-images). These include:

- `sourcegraph/postgres`
- `sourcegraph/postgres-11.1`
- `sourcegraph/redis-cache`
- `sourcegraph/redis-store`
- `sourcegraph/redis_exporter`
- `sourcegraph/zoekt-indexserver`
- `sourcegraph/zoekt-webserver`
- `sourcegraph/pgsql-exporter`
- `sourcegraph/pod-tmp-gc`
- And possibly others which we intend to open source.
