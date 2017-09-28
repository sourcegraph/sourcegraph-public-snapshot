rdeps scans GOPATH for all reverse dependencies of a set of Go
packages.

# Installation

```
go get honnef.co/go/tools/cmd/rdeps
```

# Usage

Invoke `rdeps` with zero or more arguments that are Go packages to
print their reverse dependencies.

Alternatively, use the `-stdin` flag and provide a list of Go packages
on standard input.

See `rdeps -h` for all flags.

# Example

```
$ rdeps database/sql 2>/dev/null | head -5
github.com/GoogleCloudPlatform/golang-samples/docs/appengine/cloudsql
github.com/lxc/lxd/lxd
github.com/mattn/go-sqlite3/sqltest
github.com/mgutz/dat/sql-runner
github.com/mgutz/dat
```
