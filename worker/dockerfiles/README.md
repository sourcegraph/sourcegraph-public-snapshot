srclib toolchain Docker images
==============

This directory contains Dockerfiles to generate Docker images to run
srclib toolchains inside the Drone environment the worker creates.

Instructions for deploying srclib updates to Sourcegraph.com
------------

To deploy an update to a srclib toolchain to Sourcegraph, run `make build && make push`. The changes will be reflected on the next update to Sourcegraph.com workers (the workers must be bounced so they pull the latest Docker images).

To deploy an update to srclib core to Sourcegraph, modify `SRCLIB_VERSION` in `dockerfiles/Dockerfile`. Then run `make clean && make build && make push`.
