# Sourcegraph E2E page tests

This directory contains all of the Sourcegraph E2E page tests which are tested
from an end user's perspective (i.e. from within an actual web browser) using
[Selenium](http://www.seleniumhq.org/).

- These tests should be run locally as you would any other Go tests (see `README.dev.md`).
- These tests are run automatically as part of CI.
- These tests are run automatically against sourcegraph.com and, in the event of
  a regression in the deployed version, a rollback will occur automatically.

## Adding new tests

To add a new E2E page test:

- **Important**: Ensure that your test matches the criteria listed below!
- Every test must use its own user! This contraint must be enforced to avoid
  concurrency issues. In the long run this constraint may be modified.
- Create a `mytest.go` file based on one of the existing tests in this directory
  (see `login_flow.go` for example).
- See the godoc for the [sourcegraph.com/sourcegraph/go-selenium package](https://godoc.org/sourcegraph.com/sourcegraph/go-selenium)
  which makes writing these tests very easy.

## Testing criteria

#### Do:

- Do **write E2E tests for user flows that, if broken, Sourcegraph would be
  considered very broken to end users**. Examples:
  - Registering an account.
  - Logging in as an existing user.
  - Jumping to a definition (on a prebuilt repository / one whose build can
    never break).
- Do write unit tests and integration tests for the same code that you are
  testing in E2E tests. E2E tests do not replace unit or integration tests, they
  only act as a last line of defense.

#### Don't:

- Don't use E2E tests for testing non-critical code paths (only test user flows
  that are critical to Sourcegraph users).
- Don't use E2E tests to test code that could otherwise be tested in a Go-only
  or Javascript-only unit test.
- Don't write E2E tests that can ever fail or otherwise be flaky. They must be
  _extremely reliable_ because they are constantly run against sourcegraph.com
  and will trigger an automatic rollback in the event of test failure.
