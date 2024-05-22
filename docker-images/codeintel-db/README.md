# codeintel-db

This image provides a Postresql server for the _codeintel_ database for Sourcegraph.

## Building and testing

This image comes in two flavours, the old alpine image, and the hardened one, built with Wolfi and Bazel.

### Alpine

- Build: `./docker-images/codeintel-db/build.sh`
- Test: N/A

### Hardened

Please note that migrating from the Alpine image, this require a manual step to reindex the database. See the 5.1 upgrade documentation for details.

- Build: `bazel build //docker-images/codeintel-db:image_tarball`
- Test: `bazel build //docker-images/codeintel-db:image_test`
