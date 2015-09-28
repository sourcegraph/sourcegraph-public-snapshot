# Development README (WIP)

The best way to become familiar with the Sourcegraph repository is by
reading the code at https://src.sourcegraph.com/sourcegraph.


## Code standards

The Sourcegraph repository enforces some code standards via `make
test`, which is also run in CI.

Useful reading:

* [Go "style guide"](https://github.com/golang/go/wiki/CodeReviewComments):
  we generally adhere to these guidelines. Many of them are caught by
  `gofmt`, `golint`, and `go vet` (all of which you should set up to
  run automatically in your editor). Others should be raised during
  code review.

## Quickstart

Make sure you have `node` / `npm` installed. In your terminal run:

```
make serve-dev
```

This will continuously compile your code and live reload your locally running
instance of Sourcegraph. Navigate your browser to http://localhost:3000 to
see if everything worked.

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

When in doubt, use `make test` (it always runs all tests).

Also, keep in mind that after you update dependencies, you must re-run
codegen (see the Codegen section for more info). Here's a one-liner
for updating go-sourcegraph (the most commonly updated dependency):

```
godep update sourcegraph.com/sourcegraph/go-sourcegraph/... && godep go generate ./...
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
* hackily updating protobuf-generated code to make its JSON representation CamelCase not `snake_case` (this is a tech debt TODO to obviate)
* generate wrappers for interfaces (e.g., `./server/internal/middleware/*` packages)
* pack app templates and assets into binaries

The way we use code generation is not necessarily good, so you should
think carefully before adding more.

To re-generate codegenned code, run:

```
godep go generate ./...
```

Note that you should always run this after you run `godep update` to
update dependencies.

Also, sometimes there are erroneous diffs. This occurs for a few
reasons, none of which are legitimate (i.e., they are tech debt items
we need to address):

* Your codegen tool (e.g., `go-bindata`) version might differ from
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


## Building with Docker

If [Docker](https://docs.docker.com/) is installed, you can build and
run Sourcegraph by running:

```
docker build -t src .
docker run -p 3000:3000 -p 3100:3100 src
```

Then open http://localhost:3000 to access Sourcegraph.

NOTE: Ubuntu users may need to add themselves to the docker group
(`sudo usermod -a -G docker $USER`) and restart after installing docker.
