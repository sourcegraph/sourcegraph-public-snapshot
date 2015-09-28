# srclib-sample [![Build Status](https://travis-ci.org/sourcegraph/srclib-sample.png?branch=master)](https://travis-ci.org/sourcegraph/srclib-sample)

Sample toolchain demonstrating how to build a toolchain for
[srclib](http://srclib.org) to implement source code and dependency analysis for
a programming language.

## Usage

First, install [srclib](https://srclib.org).

Then, to set up this toolchain, clone this repository. Then from the directory
you cloned it to, run:

```
src toolchain add sourcegraph.com/sourcegraph/srclib-sample
```

That adds a symlink to this directory in your SRCLIBPATH.

Now, running `src toolchain list` should show this toolchain

```
$ src toolchain list
PATH                                       TYPE
...
sourcegraph.com/sourcegraph/srclib-sample  program, docker
```

Next, build this toolchain's Docker image:

```
src toolchain build sourcegraph.com/sourcegraph/srclib-sample
```

Now try running the tests, in both variants: invoking this toolchain as a
program and invoking it in a Docker container.

```
git submodule update --init # Initializes the testing submodules.
src test -m docker
src test -m program
```
