srclib toolchain Docker images
==============

This directory contains Dockerfiles to generate Docker images for the
Sourcegraph worker.

Instructions for deploying srclib updates to Sourcegraph.com
------------

1. Push your changes to the upstream `master` of the srclib or srclib toolchain repository.
2. Run:

```
make srclib-clean
make srclib
make toolchain-repos-clean
```

If updating a single toolchain, run:

```
TOOLCHAINS=$TOOLCHAIN_NAME make build && make push
```

If updating all toolchains (or making an update to srclib core), run:

```
make build && make push
```

3. Bounce the Sourcegraph.com workers so they pick up the latest Docker images.

Development
-----------

The `Makefile` checks out copies of the srclib core and srclib
toolchain repositories to the `cache/` directory and uses these to
build the Docker images. These can be re-fetched using `make
srclib-clean && make srclib` and `make toolchain-repos-clean && make
toolchain-repos-all`.

During development of srclib, clone your local copy of srclib core or
srclib toolchain(s) to the proper subdirectory of `cache/` and then
run `TOOLCHAINS=$TOOLCHAIN_NAME make build`. Then restart your
Sourcegraph development server.
