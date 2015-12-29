# Development README

The best way to become familiar with the Sourcegraph repository is by
reading the code at https://src.sourcegraph.com/sourcegraph.

## Environment

Before you can develop Sourcegraph you'll need to set up a
development environment. Here's what you need:

- [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
- [Go](https://golang.org/doc/install) (v1.5.0 or higher)
- [node](https://nodejs.org/en/download/) (v4.0.0 or higher)
- [make](https://www.gnu.org/software/make/)
- [Docker](https://docs.docker.com/engine/installation/) (v1.8 or higher)

If you are new to Go, you should also [set up your `GOPATH`](https://golang.org/doc/code.html)
(a directory which contains all your projects).

## Get the code

Then, clone the `sourcegraph` repository into `$GOPATH/src/src.sourcegraph.com/sourcegraph`

```
go get src.sourcegraph.com/sourcegraph
```

## Build

Make sure your `$GOPATH` is set and your `$PATH` includes `$GOPATH/bin`:

```
echo $GOPATH # should print something
echo $PATH # should include $GOPATH/bin
```

Then in your terminal run:

```
cd $GOPATH/src/src.sourcegraph.com/sourcegraph
make dep
make serve-dev
```

This will continuously compile your code and live reload your locally running
instance of Sourcegraph. Navigate your browser to http://localhost:3080 to
see if everything worked.

`make serve-dev` may ask you to run ulimit to increase the maximum number
of available file descriptors for a process. You can make this setting
permanent for every shell session by adding the following line to your
`.*rc` file (usually bash or zsh):

```bash
# increase max number of file descriptors for running a sourcegraph instance.
ulimit -n 16384
```

## Test

To run all tests:

```
make test
```

To run tests within a directory (and recrusively its subdirectories):

```
godep go test ./app/...
```

To run a specific package's tests:

```
godep go test ./util/textutil
```

Learn more at [README.tests.md](README.tests.md).

## godep

The Sourcegraph repository uses
[godep](https://github.com/tools/godep) to manage Go dependencies.

During development, it's typically easiest to run `godep` directly to
build binaries and run tests. For example, you'll probably run the
following commands frequently:

```
# Compile and install a new "src" to your PATH:
godep go install ./cmd/src

# Run some tests
godep go test ./app/...

# Run some exectest-tagged tests (these tests invoke "src", so we must
# first compile the latest "src" for them to invoke):
godep go install ./cmd/src && godep go test -tags 'buildtest exectest' ./sgx
```

Also, keep in mind that after you update dependencies, you must re-run
codegen (see the Codegen section for more info). Here's a one-liner
for updating `go-sourcegraph` (the most commonly updated dependency):

```
godep update src.sourcegraph.com/sourcegraph/go-sourcegraph/... && godep go generate ./...
```

When you add a new package dependency, you need to run `godep save
./...` to add it. Note that this requires all dependencies to exist in
your system GOPATH; to make that so, run `godep restore ./...`
first. If any repos fail to clone, place them manually in the
directories underneath your GOPATH where `godep` is expecting
them. (With almost a hundred dependency repos, sometimes ephemeral
errors cause clone failures.)

Many people find `godep` confusing at first. It can be finnicky when
running `godep update` and `godep save`. Ask someone if you run into
issues with those. However, `godep go <install|test>` is quite
simple--it just runs `GOPATH=$PWD/Godeps/_workspace:$GOPATH go
<install|test>` (no other magic occurs).

## Codegen

The Sourcegraph repository relies on code generation triggered by `go
generate`. Code generation is used for a variety of tasks:

* generating code for mocking interfaces
* hackily updating protobuf-generated code to make its JSON representation CamelCase not `snake_case`
(this is a tech debt TODO to obviate)
* generate wrappers for interfaces (e.g., `./server/internal/middleware/*` packages)
* pack app templates and assets into binaries

To re-generate codegenned code, run:

```
godep go generate ./...
```

Note that you should always run this after you run `godep update` to
update dependencies.

Also, sometimes there are erroneous diffs. This occurs for a few
reasons, none of which are legitimate (i.e., they are tech debt items
we need to address):

* Your codegen tool (e.g., `gen-mocks`) version might differ from
  the last committer's in the code it produces. We need to version
  these tools.
* The codegen tool might emit code that depends on system
  configuration, such as the system timezone. We need to submit PRs to
  the tools to eliminate these issues.
* You might have existing but gitignored files that the codegen tools
  read on your disk that other developers don't have. (This occurs for
  app assets especially.)

If you think a diff is erroneous, don't commit it. Add a tech debt
item to the issue tracker and assign the person who you think is
responsible (or ask).

Finally, `go-sourcegraph` also uses many codegen tools. See its
README.md for instructions on how to install them.

## Code standards

The Sourcegraph repository enforces some code standards via `make
test`, which is also run in CI.

Read more about our style in [README.style.md](README.style.md).

## Developer API & Documentation

Sourcegraph uses gRPC and Protocol Buffers under the hood. Read more at
[README.protobuf.md](README.protobuf.md) or view the documentation for the API at [developer.sourcegraph.com](https://developer.sourcegraph.com).

## Windows notes

Windows support for Sourcegraph is currently in alpha. A few extra
steps are required to run Sourcegraph in a Windows environment:

- Sourcegraph depends on some GNU and open-source command line tools
  like `make`, `sh`, and `find`. Install Cygwin from
  http://cygwin.org/ (tested with 2.2.1 x32).
- Install Git from http://git-scm.com rather than use Cygwin's
  default Git, which does not properly handle local repositories.
- Configure your Git to use Unix-style line endings.
- Unix symlinks must be converted to Windows symlimks. This does not
  happen automatically (see
  http://stackoverflow.com/questions/5917249/git-symlinks-in-windows).
  After checking out Sourcegraph, run `bash
  dev/git-windows-aliases.sh` to install the git aliases needed to do
  the conversion. Then run `git rm-symlinks` from the repository root
  to convert all symlinks into Windows symlinks and mark them as "not
  changed" (necessary to avoid issues related to running npm-based
  tasks).

Note that multiple unit tests currently fail on Windows.
