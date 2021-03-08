# Sourcegraph dependency graph development utilities

A development utility specific to Sourcegraph package import conventions.

## Building

Run `go build` in this directory.

## Running

Run `depgraph {subcommand}` somewhere inside the sourcegraph/sourcegraph repository. This will analyze the Sourcegraph package dependency graph then perform some subcommand-specific action over it.

### Commands

The following commands are available.

#### lint

Usage: `./dev/depgraph/depgraph lint [pass...]`

This command ensures the following lint passes. Violations of the lint rules will be displayed on standard out, but the utility does not currently exit with a non-zero status.

- **NoDeadPackages**: Report unused packages (except for library code and main packages)
- **NoReachingIntoCommands**: Report packages that import code from an unrelated command
- **NoBinarySpecificSharedCode**: Report shared packages that are imported only by a single command
- **NoLooseCommands**: Report main packages outside of known command roots
- **NoUnusedSharedCommandCode**: Report packages that could be moved into an internal package
- **NoEnterpriseImportsFromOSS**: Report packages that illegally import enterprise code

#### trace

Usage: `./dev/depgraph/depgraph trace {package} [-dependencies=false] [-dependents=false]`

This command outputs a dot-formatted graph encoding the (transitive) dependencies of dependencies and (transitive) dependents of dependents rooted at the given package. Saved to a file `trace.dot`, you can convert this to SVG (or another format) via `dot -Tsvg trace.dot -o trace.svg`.
