# GraphQL integration tests

This directory contains API-based integration tests in the form of standard Go tests. It is called gqltest since most of our API is GraphQL. However, the test suite has been extended to test other endpoints such as streaming search.

To run these tests against a local change, use `sg` to spin up a job in Buildkite:

```
sg ci bazel test //testing:backend_integration_test
```

You can also push a branch to run integration tests, since they're always run by CI on branch pushes.

## How to run tests locally

You can also run integration tests locally. Note: this is not a well-trodden path, and it's recommended to use CI
instead.

Steps:

1. Request "CI Secrets Read Access" in Entitle
2. Run `sg test bazel-backend-integration`

## How to add new tests

Adding new tests to this test suite is as easy as adding a Go test, here are some general rules to follow:

- Use `gqltest-` prefix for entity name, and be as specific as possible for easier debugging, e.g. `gqltest-org-user-1`.
- Restore the previous state regardless of failures, including:
  - Delete new users created during the test.
  - Delete external services created during the test.
  - Although, sometimes you would not want to delete an entity so you could login and inspect the failure state.
