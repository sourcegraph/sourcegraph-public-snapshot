+++
title = "Failed or incomplete builds"
navtitle = "Builds"
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

## Missing Language Support

You may not have the necessary [srclib](https://srclib.org) language support toolchains installed.

* **Solution**: Run `src toolchain install LANGUAGE` (where `LANGUAGE`
  is `go` or `java`) to install the toolchains. You must run this
  command as the Sourcegraph server user.
* If you are using a language other than Go or Java, you can try
  other srclib toolchains (either from Sourcegraph or
  community-supported toolchains), but be aware that they are not
  ready for production usage.

## Non-Standard build process

Srclib toolchains generally assume the language standard for build process and directory structure. If your project uses a non-standard process, it is possible the toolchain doesn't understand your project.

* **Solution**: You can configure language toolchains to help them
  process your code by committing a special top-level file named
  `Srcfile` to your repository. Check out the READMEs for each
  toolchain to see the supported configuration settings:
  [srclib-go README](https://sourcegraph.com/sourcegraph/srclib-go)
  and
  [srclib-java README](https://sourcegraph.com/sourcegraph/srclib-java).

## Incompatible system programs/libraries

It is possible that the server you're running Sourcegraph on has system programs or libraries that are incompatible. Toolchains try to be as robust as possible to differences between operating systems, compiler/interpreter versions, etc., but sometimes there are incompatibilities. Please check the logs (see below) and report the issue.

## Insufficient RAM

If the server has insufficient RAM, builds may fail. Some large repositories require a lot of memory to compile or analyze. Check the logs (see below) and if you see out-of-memory errors, increase the amount of RAM on your server.

# Checking build logs

Full logs are kept for all builds. To view the logs, visit the detail
page for a build and click **Download logs** (for the full log) or the
**Logs** link next to the individual tasks. Include these build logs
when reporting issues.

# Contact Us

Can't figure out what is going wrong? We'd love to help! Please create an issue or [contact us](mailto:support@sourcegraph.com) and provide:

1. As much information as you can that may relate to the issue (was it just this build that failed, or all of them? etc).
2. Steps to reproduce the issue on our end, if possible.
3. The logs for the failed build (see above).
