# Continuous Integration

We have a variety of tooling on [Buildkite](https://buildkite.com/sourcegraph/sourcegraph) and [GitHub Actions](https://github.com/sourcegraph/sourcegraph/actions) for continuous integration.

- [GitHub Actions](#github-actions)
  - [Third-Party Licenses](#third-party-licenses)

## GitHub Actions

### Third-Party Licenses

We use the [`license_finder`](https://github.com/pivotal/LicenseFinder) tool to check third-party dependencies for their licenses. It runs as a [GitHub Action on pull requests](https://github.com/sourcegraph/sourcegraph/actions?query=workflow%3A%22Licenses+Check%22), which will fail if one of the following occur:

- If the license for a dependency cannot be inferred. To resolve:
  - Use `license_finder licenses add <dep> <license>` to set the license manually
- If the license for a new or updated dependency is not on the list of approved licenses. To resolve, either:
  - Remove the dependency
  - Use `license_finder ignored_dependencies add <dep> --why="Some reason"` to ignore it
  - Use `license_finder permitted_licenses add <license> --why="Some reason"` to allow the offending license

The `license_finder` tool can be installed using `gem install license_finder`. You can run the script locally using:

```sh
# updates ThirdPartyLicenses.csv
./dev/licenses.sh

# runs the same check as the one used in CI, returning status 1
# if there are any unapproved dependencies ('action items')
LICENSE_CHECK=true ./dev/licenses.sh
```

The `./dev/licenses.sh` script will also output some `license_finder` configuration for debugging purposes - this configuration is based on the `doc/dependency_decisions.yml` file, which tracks decisions made about licenses and dependencies.

For more details, refer to the [`license_finder` documentation](https://github.com/pivotal/LicenseFinder#usage).

