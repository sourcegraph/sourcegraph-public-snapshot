# E2E tests

The end-to-end tests ensure that the overall Sourcegraph user experience just works. They are fast and reliable, and they are easy to run, write, and deploy. You should write one for every important workflow.


## How do I run the end-to-end tests?

First, check that you have Docker (Docker for Mac on macOS), Python 2.7, and virtualenv installed. Then, just run `make` in this directory. If the "Screen Sharing" application on macOS prompts you for a password, it is "secret".

You can also run the tests for a specific browser, test case, or Sourcegraph instance. For example,
- If you want to run the tests against Firefox without VNC, run `make test NOVNC=1 BROWSER=firefox`.
- If you want to run only the tests that match a query string against Chrome and open up a debugger on error, run `make test OPT="--filter=$query --pause-on-error"`.
- If you want to run the tests against a local instance of Sourcegraph, run `make test SOURCEGRAPH_URL=http://localhost:3080`. Note: the local instance must be a production build and cannot be a development build. (The development build relies on a separate server to serve static assets and currently this separate server is unreachable from inside the Selenium Docker container.)

Check out the `Makefile`'s "Demo recipes" section for other ways you can run the tests.

If you're responding to a production issue, the error message in #bot-e2e will give you a one-liner to reproduce the issue.

If you're writing a test, you probably want to run the VNC viewer, the Selenium server, and the test runner separately. See the section below on writing tests for further details.


## How do I write an end-to-end test?

The only file you should have to modify is `e2etests.py` (and maybe the `Util` or `Driver` class in `e2etypes.py`). Each test is described in a `test_*` function.

To add a test:
1. Write a `test_*` function describing the desired set of user actions.
1. Add this test to the `all_tests` variable at the end of the file.

In `e2etypes.py`, you'll find utility functions and the `Driver` and `Util` classes, which provide convenience methods for finding UI components and initiating user actions.

When writing a new test, you'll want to have 3 terminal windows running the following commands:
1. `make selenium BROWSER={chrome,firefox}` -- this runs the Selenium server in a Docker container
1. `make vnc BROWSER={chrome,firefox}` -- this runs the VNC viewer
1. `make run BROWSER={chrome,firefox} SOURCEGRAPH_URL=https://sourcegraph.com OPT="--pause-on-err --filter=$MY_TEST_NAME` -- this runs the actual test

After writing the test, run `make deploy` to build, upload, and deploy the Docker image to production. In the Slack channel, you should see a goodbye message from the old test container and a hello message from the new one.


## When should I write an end-to-end test?

You should write an end-to-end test for any user workflow that is your responsibility to own. Ask yourself, "If I do not write this test, is there a non-trivial chance that someone on my team (probably me) will be woken up in the middle of the night over the next couple months by a frustrated user or customer reporting this functionality broken?" If the answer is yes, then write a test.


### Important tips

Selenium tests are easy to write, but the naked Selenium API can make it easy to write flaky tests. When writing tests, we often assume certain pre- and post-conditions hold after statements complete. However, in the world of end-to-end GUI tests, things get a little fuzzy.

For example, just because you clicked a button to trigger a menu to appear doesn't mean that the menu has appeared immediately after the button is clicked. There might be some lag time while the app waits on a request to the server.

However, if you keep just a few things in mind, you should be fine:
- **Use `wait_for` to check for pre-conditions** you expect to hold before triggering some user action.
- **Use `wait_for` to check for post-conditions** you expect to hold following some user action.
- **Wrap almost all actions in `retry`.** This will retry the action a couple of time, catching and ignoring common Selenium exceptions that do not indicate any actual bugs.
- **Never hold onto a reference to a UI element across user actions** (or other events that will change the DOM). For instance, let's say you acquire a reference to a text input and have the user type a key. If you want the user to type another key, you should re-acquire a reference to the text input instead of re-using the old reference. The old DOM element may no longer exist (this is especially true for React, which can re-render the DOM in response to any event).
- Before pushing your test, run it in a loop for both Chrome and Firefox 30 times to check if there's any flakiness.
- Use the existing tests as a model when in doubt.


## Prod

These tests run in production in much the same way as in dev. The test runner and Selenium server run in 2 separate Docker containers in the same Kubernetes pod. We have 2 pods -- one for Firefox and one for Chrome. The test runner script runs a loop over all end-to-end tests and alerts Slack and OpsGenie on error.


## How do the end-to-end tests work?

The end-to-end tests use the [Selenium Docker images](https://github.com/SeleniumHQ/docker-selenium). The test runner is a Python script and each end-to-end test is a Python function. The tests use the Selenium Remote Driver, which means the Selenium driver (and the web browser it controls) runs as a separate service from the test runner.

When running in dev, the Selenium Docker image includes a VNC server, which lets you connect a VNC client to the see exactly which actions the tests are running in the browser.


## Why can't I run the tests against a development server?

The end-to-end tests will not work against a development build because it relies on the existence of a webpack server running on localhost in the same machine as the browser. It's possible to rig this up, but it involves changing some environment variables to make it so the development server is running not on `localhost`, but on some other hostname, which will need to be mapped to the appropriate IP in the Selenium container.

Regardless of these difficulties, we should be running the end-to-end tests in an environment as close to production as possible, so it's better anyway to run them against a production build.


## Why are they written in Python?

The Python Selenium client is well-documented and widely used. This also gives us the chance to dogfood Sourcegraph on Python.


## What additional docs / references can I refer to?

If you need to look up something in the Selenium API, the official docs are here: http://selenium-python.readthedocs.io/api.html.

There are also 3rd party docs here that are easier to use, but not as comprehensive: https://seleniumhq.github.io/selenium/docs/api/py/api.html.
