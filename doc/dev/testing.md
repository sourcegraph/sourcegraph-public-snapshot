# Testing

## Go tests (backend)

To run tests for the Go backend, run `go test ./...`, or specify a package
directly, `go test ./util/textutil`.

## TypeScript tests (web app and browser extension)

- To run all unit tests, run `yarn test` from the root directory.
- To run unit tests in development (only running the tests related to uncommitted code), run `yarn test --watch`.
  - And/or use [vscode-jest](https://github.com/jest-community/vscode-jest) with `jest.autoEnable: true` (and, if you want, `jest.showCoverageOnLoad: true`)
- To debug tests in VS Code, use [vscode-jest](https://github.com/jest-community/vscode-jest) and click the **Debug** code lens next to any `test('name ...', ...)` definition in your test file (be sure to set a breakpoint or break on uncaught exceptions by clicking in the left gutter).
- You can also run `yarn test` from any of the individual project dirs (`shared/`, `web/`, `browser/`).

Usually while developing you will either have `yarn test --watch` running in a terminal or you will use vscode-jest.

### React component snapshot tests

[React component snapshot tests](https://jestjs.io/docs/en/tutorial-react) are one way of testing React components. They make it easy to see when changes to a React component result in different output. Snapshots are files at `__snapshots__/MyComponent.test.tsx.snap` relative to the component's file, and they are committed (so that you can see the changes in `git diff` or when reviewing a PR).

- See the [React component snapshot tests documentation](https://jestjs.io/docs/en/tutorial-react).
- See [existing test files that use `react-test-renderer`](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+lang:typescript+react-test-renderer) for usage examples.
- Use the jest watcher's <kbd>u</kbd> keyboard shortcut (or `yarn test --updateSnapshot`) to update all snapshot files. Be sure to review the diff!

## End-to-end (e2e) browser-based tests

E2e tests act like a user by opening a browser and clicking, typing, and navigating around in an automated fashion. They test the whole app: JS, CSS, and backend.

### Troubleshooting failing e2e tests

When an e2e test fails ([example](https://buildkite.com/sourcegraph/sourcegraph/builds/29935#1ee967cf-eb2e-4af0-8afc-0770d1779c1d)), CI displays a snapshot of the failure [inline](https://buildkite.com/docs/pipelines/links-and-images-in-log-output) in the Buildkite output and Jest prints the test name, the error, and the line of code that failed:

![image](https://user-images.githubusercontent.com/1387653/54873894-88b0c080-4d9d-11e9-9454-409ffa9864bd.png)

A video of the session is available in the **Artifacts** tab:

![image](https://user-images.githubusercontent.com/1387653/54873783-65851180-4d9b-11e9-98e5-e23b1166f2c5.png)

Here are common failure modes:

- Timed out waiting for http://localhost:7080 to be up: the `sourcegraph/server` container failed to start, so check the container logs that appear further down in the Buildkite output
- Timed out waiting for a selector to match because the CSS class in the web app changed: update the test code
- Timed out waiting for a selector to match because the page was still loading: use `waitForSelector(selector, { visible: true })`
- Page disconnected or browser session closed: another part of the test code might have called `page.close()` asynchronously, the browser crashed (check the video), or the build got canceled
- Node was detached from the DOM: the Monaco editor changes its DOM asynchronously, so wrap interactions with it in `retry()`
- `retry` is the preferred way to "poll" for a condition that cannot be expressed through `waitForSelector()` (as opposed to relying on a fixed `setTimeout()`)

Retrying the Buildkite step can help determine whether the test is flaky or broken. If it's flaky, disable it with `test.skip` and file an issue on the author.

### Running locally

To run all e2e tests locally against your dev server, **create a user `test` with password `test`, promote as site admin**, then run:

```
env GITHUB_TOKEN=<token> yarn --cwd web run test-e2e
```

> There's a test token in `../dev-private/enterprise/dev/external-services-config.json`

This will open Chromium, create an external service, clone repositories, and execute the e2e tests.

You can single-out one test with `test.only`:


```TypeScript
        test.only('widgetizes quuxinators', async () => {
            // ...
        })
```

Alternatively, you can use `-t` to filter tests: `env ... test-e2e -t "some test name"`.

### Viewing e2e tests live in CI

If CI appears stuck on e2e tests, you can view the screen in [VNC Viewer](https://www.realvnc.com/en/connect/download/viewer/) (free) by forwarding port 5900 to the pod. Find the pod name on the top right of the step in Buildkite:

![image](https://user-images.githubusercontent.com/1387653/54874133-9157c580-4da2-11e9-89c8-d8c1e53687ad.png)

> You might have to inspect element to view it.

Drop the `-N` suffix from the name, then run:

```
gcloud container clusters get-credentials ci --zone us-central1-a --project sourcegraph-dev
kubectl port-forward -n buildkite <buildkite agent pod> 5900:5900
```

Open VNC Viewer and type in `localhost:5900`. Hit <kbd>Enter</kbd> and accept the warning. Now you'll be able to see what's causing the tests to hang (e.g. a prompt or alert that hasn't been dismissed).

### Adding a new e2e test

Open `web/src/e2e/e2e.test.ts` and add a new `test`:

```TypeScript
        test('widgetizes quuxinators', async () => {
            await page.goto(baseURL + '/quuxinator/widgetize')
            await page.waitForSelector('.widgetize', { visible: true })
            // ...
        })
```

The full [Puppeteer API](https://github.com/GoogleChrome/puppeteer/blob/master/docs/api.md) is quite large, but most tests only use a few common commands:

- `await page.goto(baseURL + '/some/route')` navigate to a URL
- `await page.waitForSelector(selector, { visible: true })` wait for an element to appear
- `await page.click(selector)` click on an element (must be visible, but not necessarily in the viewport)

The easiest way to write CSS selectors is to inspect the element in your browser and look at the CSS classes. From there, you can write a selector and get immediate feedback:

![image](https://user-images.githubusercontent.com/1387653/54873834-6c605400-4d9c-11e9-823b-faa8871df395.png)

CSS selectors in e2e tests should always refer to CSS classes prefixed with `e2e-`. This makes them easy to spot in the implementation and therefor less likely to accidentally break. `e2e-` classes are never referenced in stylesheets, they are added _in addition_ to styling classes. If an element you are trying to select does not have an `e2e-` class yet, modify the implementation to add it.

If the element you are trying to select appears multiple times on the page (e.g. a button in a list) and you need to select a specific instance, you can use `data-e2e-*` attributes in the implementation:

```HTML
<div data-e2e-item-name={this.props.name}>
  <span>{this.props.name}</span>
  <button className="e2e-item-delete-button">Delete</button>
</div>
```

Then you can select the button with `[data-e2e-item-name="foo"] .e2e-item-delete-button`.

Tip: it's generally unreliable to hold references to items that are acted upon later. In other words, don't do this:

```ts
const elem = page.selector(".selector")
elem.click()
```

Do this:

```ts
page.click(".selector")
```

### E2e caveats

In the testing pyramid, e2e tests account for a small minority of all of the tests in an app. Only reach for e2e testing when it's too difficult to unit test something.

In comparison to unit tests, e2e tests are slower and flakier but often more convenient.

E2e tests are typically beneficial for testing a happy-path through a user flow, such as adding a repository then running a search.

E2e tests are probably not worth the slowness/flakiness for testing a matrix of inputs, benchmarking, checking DOM structure, or verifying the correctness of logic.

## Visual snapshot tests

### Reviewing visual changes in a PR

When you submit a PR, a check from https://percy.io/Sourcegraph/Sourcegraph will appear:

![image](https://user-images.githubusercontent.com/1387653/54873096-ac1f3f80-4d8c-11e9-93ff-377b28121df4.png)

If Percy failed CI ❌ then click on the **Details** link to review the visual changes:

![image](https://user-images.githubusercontent.com/1387653/54873144-c0177100-4d8d-11e9-839c-3e344a6d872b.jpg)

Click the image on the right to toggle between diff and full image mode to review the change. Diff mode shows the changes in red.

If the changes are intended, click **Approve** 👍

Once you approve all of the changes, the Percy check will turn green ✅

### Adding a new visual snapshot test

Open `web/src/e2e/index.e2e.test.tsx` and add a new e2e test:

```TypeScript
        test('Repositories list', async () => {
            await page.goto(baseURL + '/site-admin/repositories?query=gorilla%2Fmux')
            await page.waitForSelector('[e2e-repository-name="/github.com/gorilla/mux"]', { visible: true })
            await percySnapshot(page, 'Repositories list')
        })
```

The `percySnapshot()` function takes the snapshot and uploads it to Percy.io.

When you submit the PR, Percy will fail until you approve the new snapshot.

### Flakiness in snapshot tests

Flakiness in snapshot tests can be caused by the search response time, order of results, animations, premature snapshots while the page is still loading, etc.

This can be solved with [Percy specific CSS](https://docs.percy.io/docs/percy-specific-css) that will be applied only when taking the snapshot and allow you to hide flaky elements with `display: none`.

## Continuous Integration

The test suite is exercised on every pull request. For the moment CI output
access is limited to Sourcegraph employees, though we hope to enable public
read-only access soon.

The test pipeline is generated by `dev/ci/gen-pipeline.go`, and written to a
YAML file by dev/ci/init-pipeline.yml. This pipeline is immediately scheduled by
Buildkite to run on the Sourcegraph build farm. Some things that are tested
include:

- all of the Go source files that have tests
- dev/check/all.sh (gofmt, lint, go generator, no Security TODOs, Bash syntax, others)
- JS formatting/linting (prettier, tslint, stylelint, graphql-lint)
- Dockerfile linter (hadolint)
- Check whether the Go module folders are "tidy" (go mod tidy)

## Release testing

To manually test against a Kubernetes cluster, use https://k8s.sgdev.org.

For testing with a single Docker image, run something like
```
IMAGE=sourcegraph/server:3.10.4 ./dev/run-server-image.sh
```
