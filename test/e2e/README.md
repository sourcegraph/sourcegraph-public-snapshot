# Sourcegraph E2E page tests

This directory contains all of the Sourcegraph E2E page tests which are tested
from an end user's perspective (i.e. from within an actual web browser) using
[Selenium](http://www.seleniumhq.org/).

- These tests should be run locally as you would any other Go tests (see `README.dev.md`).
- These tests are run automatically against sourcegraph.com and, in the event of
  a regression in the deployed version, a rollback will occur automatically.

## Running locally

End-to-end tests are run constantly against sourcegraph.com.
You may additionally run them locally on your machine:

1. Run Selenium server:
    - `docker run -d -p 4444:4444 selenium/standalone-chrome:2.52.0` is via
      docker, and is headless (ie you can't watch the run)
    - `brew install Caskroom/cask/java selenium-server-standalone
      chromedriver` via brew to run tests in a local browser. Then run
      `selenium-server -p 4444`
2. Run an E2E tests:
    - `SELENIUM_SERVER_IP=<DOCKER_HOST_IP> TARGET=https://sourcegraph.com go test -run TestDefFlow`
        - OS X: use `docker-machine ls` to find the IP of the machine if using docker.
        - Other: Just use `localhost`.

Alternatively you can use the full e2etest stack that runs in production:
1. Run the tests
    - `go install sourcegraph.com/sourcegraph/infrastructure/docker-images/e2etest`
    - `SELENIUM_SERVER_IP=<DOCKER_HOST_IP> TARGET=https://sourcegraph.com e2etest -once`
        - OS X: use `docker-machine ls` to find the IP of the machine.
        - Linux: Just use `localhost`.
2. Run a specific E2E test once: `e2etest -once -run="login_flow"`
3. Run tests against local Sourcegraph instance: specify `TARGET=http://<LOCAL_MACHINE_IP>:3080` NOT `localhost` (Selenium runs inside a Docker container, use LAN IP instead).

For authentication with the `TARGET` server, your `~/.sourcegraph/id.pem` is used by default. Set `ID_KEY_DATA=...` to specify a Base64-encoded form of this file.

## Adding new tests

To add a new E2E page test:

- **Important**: Ensure that your test matches the criteria listed below!
- Every test must use its own user! This contraint must be enforced to avoid
  concurrency issues. In the long run this constraint may be modified.
- Create a `mytest.go` file based on one of the existing tests in this directory
  (see `login_flow.go` for example).
- See the godoc for the [sourcegraph.com/sourcegraph/go-selenium package](https://godoc.org/sourcegraph.com/sourcegraph/go-selenium)
  which makes writing these tests very easy.
- Don't use `t.Fatalf` as assertions, rather rely on `WaitForCondition`. It
  leads to more reliable E2E tests.

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
