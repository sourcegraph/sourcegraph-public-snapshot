# makex

**makex** is a `make` clone for Go that makes it easier to write build tools in Go. It lets you define tasks and dependencies in the familiar Makefile format, and unlike just shelling out to `make`, it gives you programmatic access (in Go) to the progress and console output of your tasks. We use makex at [Sourcegraph](https://sourcegraph.com) to compile and analyze hundreds of thousands of your open-source projects.

Documentation: [makex on Sourcegraph](https://sourcegraph.com/sourcegraph.com/sourcegraph/makex)

[![Build Status](https://travis-ci.org/sourcegraph/makex.png?branch=master)](https://travis-ci.org/sourcegraph/makex)
[![status](https://sourcegraph.com/api/repos/sourcegraph.com/sourcegraph/makex/badges/status.png)](https://sourcegraph.com/sourcegraph.com/sourcegraph/makex)

## Install

```
go get sourcegraph.com/sourcegraph/makex/cmd/makex
makex -h
```

## Usage

Try running makex on the Makefiles in the `testdata/` directory.

```bash
$ cat testdata/sample0/y
cat: testdata/sample0/y: No such file or directory
$ makex -v -C testdata -f testdata/Makefile.sample0
makex: [sample0/y] mkdir -p sample0
makex: [sample0/y] echo hello bar > sample0/y
$ cat testdata/sample0/y
hello bar
$ makex -v -C testdata -f testdata/Makefile.sample0
$
```

```bash
$ ls testdata/sample1/
y1
$ makex -v -C testdata -f testdata/Makefile.sample1
makex: [sample1/y0] echo hello bar > sample1/y0
$ ls testdata/sample1/
y0  y1
$ makex -v -C testdata -f testdata/Makefile.sample1
$
```

## Known issues

makex is very incomplete.

* No support for setting or expanding variables (except for special-cased support for `$@` and `$^`)
* No support for filesystem globs except in the OS filesystem (not in VFS filesystems).
* Many other issues.

