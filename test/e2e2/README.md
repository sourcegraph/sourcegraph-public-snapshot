# E2E tests

The end-to-end tests are fast, reliable, and easy to write and run. They ensure that the overall Sourcegraph experience just works for users, so you should write one for every important workflow.

## How do I run the end-to-end tests in dev?

It's easy--just run `make`. (You need Docker, Python 2.7, and virtualenv installed first.) If the "Screen Sharing" application prompts you for a password, it is "secret".

- If you want to watch the show, run `make TV=1`.
- If you want to run a subset of tests that match some query string, run `make TEST=$query`.
- If you want to run tests for just Chrome or Firefox, run `make TEST={chrome,firefox}`.
- If you want the tests to pause on an error, run `make TV=1 OPT="--pause-on-err"`.


## When should I write an end-to-end test?

You should write an end-to-end test for any user workflow that is your responsibility to own. Ask yourself, "If I do not write this test, is there a non-trivial chance that someone on my team (probably me) will be woken up in the middle of the night over the next couple months by a frustrated user or customer reporting this functionality broken?" If the answer is yes, then consider writing an end-to-end test.


## How do I write an end-to-end test?

The only file you should have to modify is `e2etests.py`. Each test is described in a `test_*` function.

To add a test:
- Write a `test_*` function describing the desired set of user actions.
- Add this test to the `all_tests` variable at the end of the file.

In `e2etypes.py`, you'll find utility functions and a `Driver` class that provides convenience methods for finding UI components and initiating user actions.


### Important tips

Selenium tests are easy to write, but the naked Selenium API can make it easy to write flaky tests. When writing tests, we often assume certain pre- and post-conditions hold after statements complete. However, in the world of end-to-end full-GUI tests, things get a little fuzzy. Just because you clicked a button to trigger a menu to appear doesn't mean that the menu has appeared immediately after the button is clicked. There might be some lag time while the app waits on a request to the server.

However, if you keep just a few things in mind, you should be fine:
- **Use `wait_for` to check for pre-conditions** you expect to hold before triggering some user action.
- **Use `wait_for` to check for post-conditions** you expect to hold following some user action.
- **Wrap almost all actions in `retry`.** This will retry the action a couple of time, catching and ignoring common Selenium exceptions that do not indicate any actual bugs.
- **Never hold onto a reference to a UI element across user actions** (or other events that will change the DOM). For instance, let's say you acquire a reference to a text input and have the user type a key. If you want the user to type another key, you should re-acquire a reference to the text input instead of re-using the old reference. The old DOM element may no longer exist (this is especially true for React, which can re-render the DOM in response to any event).
- Before pushing your test, use `make test-30 TEST=$YOUR_TEST` to run it in a loop on both Chrome and Firefox 30 times to check if there's any flakiness. (Note: the Selenium Chrome driver tends to run quicker and the Firefox driver tends to be a little flakier.)
- Use the existing tests as a model when in doubt.


## Prod

These tests run in production in much the same way as in dev. The test runner script and the 2 Selenium servers are deployed in separate Docker containers in the same Kubernetes pod. The test runner script runs a loop over all end-to-end tests and alerts Slack and OpsGenie on error. In addition, the test runner container heartbeat pings OpsGenie, which will alert if the test runner ever goes down.


## How do the end-to-end tests work?

The end-to-end tests use the [Selenium Docker images](https://github.com/SeleniumHQ/docker-selenium). The test runner is a Python script and each end-to-end test is a Python function.

They use the Selenium Remote Driver, which means the Selenium driver and web browser run as a separate service from the test runner.

When running in dev, the Selenium Docker image includes a VNC server, which lets you, the developer, connect a VNC client to the see exactly which actions the tests are running in the browser.
