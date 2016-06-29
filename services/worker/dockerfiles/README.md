srclib toolchain Docker images
==============

This directory contains Dockerfiles to generate Docker images for the
Sourcegraph worker.

Instructions for deploying srclib updates to Sourcegraph.com
------------

0. Push your changes to the upstream `master` of the srclib or srclib toolchain repository.
0. Run:

```
make clean && make srclib
make build
```

If updating a single toolchain, run:

```
TOOLCHAINS=$TOOLCHAIN_NAME make build
```

0. Update `srclib_images.go`.
0. Push the new version of Sourcegraph.

Development
-----------

### srclib core
During development of srclib core, run `DEV=1 make clean && make srclib && make build`

### srclib toolchain
During development of a srclib toolchain, clone a local copy of your toolchain
repository to a subdirectory and pass it to the Dockerfile:

```
docker build --build-arg TOOLCHAIN_URL=path/to/local/toolchain/repo -t local-toolchain -f ./Dockerfile.srclib-${language-name} .
```

Then update `srclib_images.go` to set the toolchain image to be "local-toolchain".

You'll need to restart your Sourcegraph server for the new Docker image to be picked up.
