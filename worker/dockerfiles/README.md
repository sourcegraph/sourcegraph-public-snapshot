srclib toolchain Docker images
==============

This directory contains Dockerfiles to generate Docker images to run
srclib toolchains inside the Drone environment the worker creates.

Instructions for deploying srclib updates to Sourcegraph.com
------------

To deploy an update to a srclib toolchain to Sourcegraph, run `make build && make push`. The changes will be reflected on the next update to Sourcegraph.com workers (the workers must be bounced so they pull the latest Docker images).

To deploy an update to srclib core to Sourcegraph, run `make clean && make build && make push`.

dev
-----------
If you are working on toolchain srclib-LANG, in order to test your changes
with the `src` you can do the following:

- set TOOLCHAIN_URL environment variable that points to local directory 
- make changes in your code
- run `make srclib-LANG`

You may also set TOOLCHAIN_URL to HTTP, SSH, or Git URL if you need to build Docker image for not-standard toolchain repository.

The same way, by setting SRCLIB_URL environment variable you may control origin of `srclib` binary 
