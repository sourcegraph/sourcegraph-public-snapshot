# Sourcegraph dependency graph development utilities

A development utility specific to Sourcegraph package import conventions.

## Building

Run `go build` in this directory.

## Running

Run `./dev/depgraph/depgraph {subcommand}` to analyze the Sourcegraph package dependency graph. The analyzed package dependency graph will be cached in a file `depgraph.cache`. Remove this file to trigger a fresh package analysis.

This process is pretty slow - someone please fix it :)

### Commands

The following commands are available

#### clear-cache

Remove the cache file.

#### lint

Usage: `./dev/depgraph/depgraph lint [pass...]

This command ensures the following lint passes. Violations of the lint rules will be displayed on standard out, but the utility does not currently exit with a non-zero status.

- **NoDeadPackages**: Report packages with no users (except for cmd roots)
- **NoReachingIntoCommands**: Report packages that import from a command they are not a part of
- **NoSingleDependents**: Report packages that are imported by only one unrelated dependent

#### trace

Usage: `./dev/depgraph/depgraph trace {package} [-dependencies=false] [-dependents=false]`

This command outputs a dot-formatted graph encoding the (transitive) dependencies of dependencies and (transitive) dependents of dependents rooted at the given package. Saved to a file `trace.dot`, you can convert this to SVG (or another format) via `dot -Tsvg trace.dot -o trace.svg`.
