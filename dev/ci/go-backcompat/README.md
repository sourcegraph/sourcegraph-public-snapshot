The `flakefiles` directory contains _flakefiles_ that lists tests to disable when running Postgres backwards compatibility tests (see [go-backcompat/test.sh](../go-backcompat/test.sh)).

Since we are running _historic_ tests, we are unable to change or fix them in the presence of flake or acceptable/spurious failures. In these cases, we can add the tests to an explicit lists of tests that we disable prior to invoking `go test`. The content of each flakefile looks similar to the following:

```json
[
  { "path": "cmd/A/server", "prefix": "TestA", "reason": "Unused outside of Cloud." },
  { "path": "cmd/B/server", "prefix": "TestB", "reason": "Test was determiened to be flaky." },
  { "path": "cmd/C/server", "prefix": "TestC", "reason": "Dropped column presenting security issue." }
]
```

Each line indicates a _path_ and _prefix_ pair, where any test starting with the given prefix in a Go test file under the given path will be disabled (renamed so that it's invoked on the following `go test` invocations). This is limited to top-level tests defined as functions in a test package (and not sub-tests invoked by `t.Run(...)`.

Each flakefile should have a name of the form `v3.{minor}.0.json`, where the version indicates the time at which tests should be disabled. Note that the most recent release will only ever be relevant day-to-day.
