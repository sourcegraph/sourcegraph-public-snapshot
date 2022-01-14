# Testing principles

This file documents how we test code at Sourcegraph.

Related pages: [How to run tests](../how-to/testing.md) | [Testing Go code](languages/testing_go_code.md) | [Testing web code](testing_web_code.md) | [Continuous integration](continuous_integration.md)

## Philosophy

We rely on automated testing to ensure the quality of our product.

Any addition or change to our codebase should be covered by an appropriate amount of automated tests to ensure that:

1. Our product and code works as intended when we ship it to customers.
1. Our product and code doesn't accidentally break as we make changes over time.

A good automated test suite increases the velocity of our team because it allows engineers to confidently edit and refactor code, especially code authored by someone else.

Engineers should budget an appropriate amount of time for writing tests when making iteration plans.

## Testing code

In order to ensure we are true to our [philosphy](#philosophy), we have various implementations of testing for our code base. 

This includes, but is not limited to:
- Image Vulnerability scanning
- Infrascture as code 
- Unit, Integration and end-to-end tests as outlined in the [testing-pyrmid](#testing-pyramid)

Our goal is to ensure that our product and code work, and that all reasonable effort has been taken to reduce the risk of a security-related incident associated to Sourcegraph.

## Flaky tests

A *flaky* test is defined as a test that is unreliable or non-deterministic, i.e. it exhibits both a passing and a failing result with the same code.

Typical reasons why a test may be flaky:

- Race conditions or timing issues
- Caching or inconsistent state between tests
- Unreliable test infrastructure (such as CI)
- Reliance on third-party services that are inconsistent

**We do not tolerate flaky tests of any kind.** Any engineer that sees a flaky test in [continuous integration](./continuous_integration.md) should immediately:

1. Open a PR to disable the flaky test.
1. Open an issue to re-enable the flaky test (use the [Flaky Test template](https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=&template=flaky_test.md&title=Flake%3A+%24TEST_NAME+disabled)), and assign it to the most likely owner, and add it to the current release milestone.

If the build or test infrastructure itself is flaky, then [open an issue](https://github.com/sourcegraph/sourcegraph/issues/new?labels=team/distribution) and notify the [distribution team](https://handbook.sourcegraph.com/engineering/distribution#contact).

Why are flaky tests undesirable? Because these tests stop being an informative signal that the engineering team can rely on, and if we keep them around then we eventually train ourselves to ignore them and become blind to their results. This can hide real problems under the cover of flakiness.

## Broken builds on the `main` branch

A red `main` build is not okay and must be fixed. Consecutive failed builds on the `main` branch means that [the releasability contract is broken](https://handbook.sourcegraph.com/engineering/continuous_releasability#continuous-releasability-contract), and that we cannot confidently ship that revision to our customers nor have it deployed in the Cloud environment.

### Process

> In essence: Someone must have eyes on the build failure. Unsure about what's happening? Get help on #buildkite-main.

- When a PR breaks the build, the author is responsible for investigating why, and asking for help if necessary:
  - The failure will appear on [#buildkite-main](https://sourcegraph.slack.com/archives/C02FLQDD3TQ).
  - If you've done ~30 mins of investigation and the cause is still unclear, ask for help!
  - Handing the issue over to someone else (for any reason) is totally okay, but it has to happen.
  - If there's no action being taken after a reasonable amount of time, the offending PR can be reverted by anyone blocked by it.
- If there is reasonable suspicion of a [flake](#flaky-tests) (e.g. can't reproduce the problem locally) or if it’s clear that the cause is not related to the PR:
  - Rebuild the job.
  - Notify the team in charge of the concerned test or disable it.
  - It's a CI flake? Pass ownership to the DX team.
- If there is no immediate fix in sight (or rebuilding didn't fix it):
  - [Mark the faulty test as skipped or revert the changes](#flaky-tests) to restore the main branch to green and avoid blocking others.
  - if reverting won't fix because it depends on external resources, just comment out that test and open a ticket mentioning the owners.

## Testing pyramid

![Testing pyramid](testing-pyramid.svg)

### Unit tests

Unit tests test individual functions in our codebase and are the most desirable kind of test to write.

Benefits:

- They are usually very fast to execute because slow operations can be mocked.
- They are the easiest tests to write, debug, and maintain because the code under test is small.
- They only need to run on changes that touch code which could make the test fail, which makes CI faster and minimizes the impact of any [flakiness](#flaky-tests).

Tradeoffs:

- They don't verify our systems are wired up correctly end-to-end.

### Integration tests

Integration tests test the behavior of a subset of our entire system to ensure that subset of our system is wired up correctly.

Benefits:

- To the extent that fewer systems are under test compared to e2e tests, they are faster to run, easier to debug, have clearer ownership, and less vulnerable to [flakiness](#flaky-tests).
- They only need to run on changes that touch code which could make the test fail, which makes CI faster and minimizes the impact of any [flakiness](#flaky-tests).

Tradeoffs:

- They don't verify our systems are wired up correctly end-to-end.
- They are not as easy to write as unit tests.

Examples:

- Tests that call our search API to test the behavior of our entire search system.
- Tests that validate UI behavior in the browser while mocking out all network requests so no backend is required.

### End-to-end tests (e2e)

E2e tests test our entire product from the perspective of a user. We try to use them sparingly. Instead, we prefer to get as much confidence as possible from our [unit tests](#unit-tests) and [integration tests](#integration-tests).

Benefits:

- They verify our systems are wired up correctly end-to-end.

Tradeoffs:

- They are typically the slowest tests to execute because we have to build and run our entire product.
- They are the hardest tests to debug because failures can be caused by a defect anywhere in our system. This can also make ownership of failures unclear.
- They are the most vulnerable to [flakiness](#flaky-tests) because there are a lot of moving parts.

Examples:

- Run our Sourcegraph Docker image and verify that site admins can complete the registration flow.
- Run our Sourcegraph Docker image and verify that users can sign in and perform a search.

### Visual testing

Visual testing is useful to catch visual regressions and verify designs for new features. [More info about visual testing philosophy](testing_web_code.md#visual-regressions)

We use [Chromatic Storybook](https://www.chromatic.com/) to detect visual changes in specific React components. Post a message in #dev-chat that you need access to Chromatic, and someone will add you to our organization (you will also receive an invitation via e-mail). You should sign into Chromatic with your GitHub account. If a PR you author has visual changes, a UI Review in Chromatic will be generated. It is recommended that a designer approves the UI review.

We use [Percy](https://percy.io/) to detect visual changes in Sourcegraph features during browser-based tests (client integration tests and end-to-end tests). You may need permissions to update screenshots if your feature introduces visual changes. Post a message in #dev-chat that you need access to Percy, and someone will add you to our organization (you will also receive an invitation via e-mail). Once you've been invited to the Sourcegraph organization and created a Percy account, you should then link it to your GitHub account.

## Ownership

- [DevX Team](https://handbook.sourcegraph.com/engineering/enablement/dev-experience) owns build and test infrastructure.
- [Frontend Platform Team](https://handbook.sourcegraph.com/engineering/enablement/frontend-platform) owns any tests that are driven through the browser.

## Conventions

- **Naming tests in Go code.** We strive to follow the same naming convention for Go test functions as described for [naming example functions in the Go testing package](https://golang.org/pkg/testing/#hdr-Examples).
