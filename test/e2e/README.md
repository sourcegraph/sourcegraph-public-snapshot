# Sourcegraph E2E tests

E2E (end-to-end) tests are page tests running from an end user's perspective
(within a real web browser) using [Selenium](http://www.seleniumhq.org/). They
are ran automatically 24/7 against Sourcegraph.com and automatically trigger
a version rollback in the event of a regression.

- [Installation](#installation)
- [Testing](#testing)
- [Additional options](#additional-options)
- [Creating tests](#creating-tests)
- [Testing criteria](#testing-criteria)

## Installation

**OS X** (testing in Google Chrome)

```bash
brew install Caskroom/cask/java selenium-server-standalone chromedriver
```

**Any OS** (headless testing)

```bash
docker run -d -p 4444:4444 selenium/standalone-chrome:2.52.0
```

Note: Some Docker versions may need you to set the `SELENIUM_SERVER_IP`
environment variable to the one listed by `docker-machine ls`, and use
`TARGET=https://<DOCKER_IP>:3080` instead of localhost when testing. With
"Docker For Mac", however, this is not needed.

# Testing

First `cd test/e2e`, then:

```bash
TARGET=http://localhost:3080 go test
```

## Common problems

* If the tests are flaky and you are seeing messages like `Exception:
  chrome not reachable` in the Selenium logs, run with the `go test
  -parallel 1` flag.


## Additional options

See environment variables [in the source](https://sourcegraph.com/sourcegraph/sourcegraph/-/blob/test/e2e/e2etest.go#L797-813).

## Creating tests

To add a new E2E page test:

  1. **Important: Ensure that your test matches the testing criteria below!**
  2. Every test must use its own user!** This contraint must be enforced to avoid
  any concurrency issues.
  3. Base your test on an existing one (see `repo_flow.go` for example)
  4. Add a new `TestFoo` entry to `e2etest_test.go`.
  5. See the [test/e2e API](https://godoc.org/sourcegraph.com/sourcegraph/sourcegraph/test/e2e)
     and the [go-selenium API](https://godoc.org/sourcegraph.com/sourcegraph/go-selenium) docs.

General wisdom:

  - Always prefer the helpers in the test/e2e API over methods provided by go-selenium,
    because they are always more reliable.
  - Always use `WaitForCondition` over `t.Fatalf` for assertions, as it is more reliable.

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
