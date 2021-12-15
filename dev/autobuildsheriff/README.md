# autobuildsheriff

`autobuildsheriff` is designed to be run on a cron to respond to periods of consecutive build failures on a Buildkite pipeline.
Owned by the DevX team.

## Testing

- `branch_test.go` contains integration tests against the GitHub API. Normally runs against recordings in `testdata` - to update `testdata`, run the tests with the `-update` flag.
- All other tests are strictly unit tests.
