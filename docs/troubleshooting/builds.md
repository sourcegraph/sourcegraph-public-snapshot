+++
title = "Failed or incomplete code analysis"
linktitle = "Code Intelligence"
+++

In Sourcegraph, a "build" is the analysis process of an entire
commit's worth of code. The analysis consists of
[srclib](https://srclib.org) language support toolchains parsing,
compiling, and semantically analyzing your code. The output is a set
of data files that describe every definition, reference, and other
code objects, as well as indexes of this data.

# Common Build Failures

Building is a complex process, and sometimes builds fail or lack 100%
coverage of your code. The most common reasons for builds to fail or
be incomplete are:

## Non-Standard build process

Srclib toolchains generally assume the language standard for build process and directory structure. If your project uses a non-standard process, it is possible the toolchain doesn't understand your project.

* **Solution**: You can configure language toolchains to help them
  process your code by committing a special top-level file named
  `Srcfile` to your repository. Check out the READMEs for each
  toolchain to see the supported configuration settings:
  [srclib-go README](https://sourcegraph.com/sourcegraph/srclib-go)
  and
  [srclib-java README](https://sourcegraph.com/sourcegraph/srclib-java).

## Insufficient RAM

If the server has insufficient RAM, builds may fail. Some large repositories require a lot of memory to compile or analyze. Check the logs (see below) and if you see out-of-memory errors, increase the amount of RAM on your server.

# Checking build logs

Full logs are kept for all builds. To view the logs, visit the detail
page for a build and click **Download logs** (for the full log) or the
**Logs** link next to the individual tasks. Include these build logs
when reporting issues.

# Contact Us

Can't figure out what is going wrong? We'd love to help! Please create an issue or [contact us](mailto:help@sourcegraph.com) and provide:

1. As much information as you can that may relate to the issue (was it just this build that failed, or all of them? etc).
2. Steps to reproduce the issue on our end, if possible.
3. The logs for the failed build (see above).
