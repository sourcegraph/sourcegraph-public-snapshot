# buildchecker

`buildchecker` is designed to respond to periods of consecutive build failures on a Buildkite pipeline.
Owned by the [DevX team](https://handbook.sourcegraph.com/departments/product-engineering/engineering/enablement/dev-experience).

More documentation for Sourcegraph teammates is available in the[CI incidents playbook](https://handbook.sourcegraph.com/departments/product-engineering/engineering/process/incidents/playbooks/ci#scenarios).

## Usage

```sh
go run ./dev/buildchecker/ # directly
./dev/buildchecker/run.sh  # using wrapper script
```

Also see the [`buildchecker` GitHub Action workflow](../../.github/workflows/buildchecker.yml) where this program is run on an automated basis.

## Development

- `branch_test.go` contains integration tests against the GitHub API. Normally runs against recordings in `testdata` - to update `testdata`, run the tests with the `-update` flag.
- All other tests are strictly unit tests.
