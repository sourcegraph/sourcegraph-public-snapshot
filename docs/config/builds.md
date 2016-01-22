+++
title = "Builds"
linktitle = "Builds"
+++

In Sourcegraph, a "build" runs the the analysis process of an entire
commit's worth of code. The analysis consists of
[srclib](https://srclib.org) language toolchains parsing,
compiling, and semantically analyzing your code. The output is a set
of data files that describe every definition, reference, and other
code objects, as well as indices of this data for quick lookup.

A build process is triggered automatically whenever you push code
or tags to your repository. The build process is managed by
[Drone](https://github.com/drone/drone), an award-winning open-source
continuous integration system built on Docker which comes bundled
natively with Sourcegraph.

Drone can also run your tests. Check out the [getting started](http://readme.drone.io/usage/overview/)
guide for instructions on how to configure your project.

# Common build failures

Building is a complex process, and sometimes analysis fails or lacks 100%
coverage of your code. The most common reasons for code analysis to fail or
be incomplete are:

## Non-standard build process

Srclib toolchains generally assume the language standard for build process and directory structure. If your project uses a non-standard process, it is possible the toolchain doesn't understand your project.

* **Solution 1**: You can configure language toolchains to help them
  process your code by committing a special top-level file named
  `Srcfile` to your repository. Check out the READMEs for each
  toolchain to see the supported configuration settings:
  [srclib-go README](https://sourcegraph.com/sourcegraph/srclib-go)
  and
  [srclib-java README](https://sourcegraph.com/sourcegraph/srclib-java).
* **Solution 2**: You can provide custom build configuration to Drone in a `.drone.yml`. See the
  [getting started](http://readme.drone.io/usage/overview/) guide.

## Insufficient RAM

If your server has insufficient RAM, builds may fail. Some large repositories require a lot of memory to compile or analyze. Check the logs (see below) and if you see out-of-memory errors, increase the amount of RAM on your server.

# Turning off builds

To stop Sourcegraph from running builds, run your server with the `--no worker` flag.

```
src serve --no-worker
```

Or if you run Sourcegraph with a daemon, add the following to your `/etc/sourcegraph/config.ini`:

```
[serve]
no-worker = true
```

# Contact Us

Have a problem? We'd love to help! Please
[create an issue](https://src.sourcegraph.com/sourcegraph/.tracker/new) in our Tracker or
[contact us](mailto:help@sourcegraph.com) and provide:

1. As much information as you can that may relate to the issue:
  - was it a single build that failed, or all of them
  - which step of the build failed
  - details about your repository (language, etc)
  - build logs
2. Steps to reproduce the issue on our end, if possible.
