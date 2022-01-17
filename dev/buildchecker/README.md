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

### History

```sh
go run ./dev/buildchecker -buildkite.token=$BUILDKITE_TOKEN -failures.timeout=999 -created.from="2021-08-01" history
```

Writes aggregated data, including the builds it finds, to a few files. To load builds from a file instead of fetching from Buildkite, use `-load-from="$FILE"`.

## Development

- `branch_test.go` contains integration tests against the GitHub API. Normally runs against recordings in `testdata` - to update `testdata`, run the tests with the `-update` flag.
- All other tests are strictly unit tests.
