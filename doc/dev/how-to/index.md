# How to guides

## Documentation

- [How to write great docs](https://documentation.divio.com/) (watch the video)

## New features

- [How to add support for a language](add_support_for_a_language.md)
- [Release Browser Extensions](releasing_browser_extensions.md)

## Implementing Sourcegraph

- [Developing the product documentation](documentation_implementation.md)
- [Observability](../background-information/observability/index.md)
  - [How to find monitoring](find_monitoring.md)
  - [How to add monitoring](add_monitoring.md)

## Local Environment

- [Set up local development with Zoekt and Sourcegraph](zoekt_local_dev.md)
- [How to debug live code](debug_live_code.md)
- [Ignoring editor config files in Git](ignoring_editor_config_files.md)
- [Use `golangci-lint`](use-golangci-lint.md)

## Testing Sourcegraph & CI

- [How to run tests](testing.md)
   - See also [Testing Principles](../background-information/testing_principles.md) and [Continuous Integration](../background-information/continuous_integration.md)
- [Configure a test instance of Phabricator and Gitolite](configure_phabricator_gitolite.md)
- [Test a Phabricator and Gitolite instance](test_phabricator.md)
- [Adding or changing Buildkite secrets](adding_buildkite_secrets.md)

## Executors

- [How to deploy a new executor image](deploy_executor_image.md)

## Windows support

Running Sourcegraph on Windows is not actively tested, but should be possible within the Windows Subsystem for Linux (WSL).
Sourcegraph currently relies on Unix specifics in several places, which makes it currently not possible to run Sourcegraph directly inside Windows without WSL.
We are happy to accept contributions here! :)

