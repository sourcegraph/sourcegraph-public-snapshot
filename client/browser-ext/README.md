# Sourcegraph browser extensions for Google Chrome and Firefox

## Project structure

```
browser-ext
├── app <-- [React](https://facebook.github.io/react/) + [Redux](http://redux.js.org/) application
│	└── actions <-- methods to fetch from sourcegraph API & change state
│	└── analytics <-- user event logger
│	└── components <-- receive state from reducers as and re-render whenever a property it subscribes to is updated
│	└── constants <-- the names given to actions
│	└── reducers <-- "holder" of current state: functions change current state when actions are dispatched
│	└── store  <-- "persistence of current state
│	└── utils  <-- logic to apply annotations to a blob, misc. utility helpers
├── chrome
│	└── assets <-- an icon for the Chrome/Firefox store
│	└── extension
│		└── background
│			└── inject.js <-- for development only (hot reloading)
│			└── storage.js <-- wrapper for chrome.storage get/set; necessary due to differences in
│								Firefox/Chrome security models
│		└── background.js <-- loads scripts in ./background
│		└── inject.js <-- injects app/components onto the page
│	└── views <- these are just dumb script holders; jade templating is used to get build-time information
│				(environment) into the extension
│	└── manifest.prod.json <-- explains to Chrome/Firefox how to load the extension and what permissions it needs
│	└── manifest.dev.json <-- dev version of ^^
├── scripts <-- build/development scripts
├── webpack <-- build configuration
```

## Architecture

The browser (Firefox/Chrome) will load a script onto the page when the user
visits GitHub.com or Sourcegraph.com.

The script (injects.js) will inject application components ("modules")
which may either have UI (as in normal React) or only have side effects
(e.g. the "BlobAnnotator" doesn't render itself, but is responsible for
updating the GitHub page dom to include tooltips/links).

In both cases, we do this using React.Component as the module and
rely on its normal lifecycle methods for most of the control flow.
The state container is vanilla Redux, and any component/module can subscribe
to any property on the reducer state to go through a re-render cycle when that
property changes.

Actions are provided on a Component's this.props via the @connect decorator.
Use these to make API requests and update application state.

## Requirements

- `npm` >= 3.6.0
- `node` >= 5.6.0
- `yarn` >= 0.16.0
- `make`

The latest stable version of node will suffice, which you can install as follows after you have installed npm:

```
sudo npm cache clean -f
sudo npm install -g n
sudo n stable

sudo ln -sf /usr/local/n/versions/node/<VERSION>/bin/node /usr/bin/node
```

## Installation

```bash
$ npm install -g yarn
$ yarn install
```

## Development

```bash
$ yarn run dev
```
* Allow `https://localhost:3000` (insecure) connections in Chrome (navigate to https://localhost:3000, click "ADVANCED",
then "Proceed to localhost"). This is necessary because pages are injected on Sourcegraph/GitHub (https), so `webpack-dev-server`
procotol must also be https.
* [Load unpacked extensions](https://developer.chrome.com/extensions/getstarted#unpacked) with `./dev` folder.
* Webpack will manage hot reloading via `react-transform`.

#### Using Redux DevTools Extension

You can use [redux-devtools-extension](https://github.com/zalmoxisus/redux-devtools-extension) in development mode.
You can also uncomment the code in configureStore.dev.js to have action/state change logging in the
dev console, though it can get a little verbose.

## Build

```bash
$ make build # create unzipped distribution artifact in ./build; required for e2e tests
```

## Test

```bash
$ make test-unit # run unit tests
$ make test-watch # watch for changes & run unit tests
$ make test-e2e # run e2e tests
$ make test-all # run all tests
```

End-to-end tests for the extension are located above this project root, at `test/e2e2`.

## Create Distributions

```bash
$ make dist
```

This make target process does more than just compress the build directory, to ensure the environment used to produce build artifacts
mimics the environment for the extension submission teams.

## Deploy (chrome extension only)

```bash
$ make deploy
```

To deploy the chrome extension with your Google Apps credentials, you must have `CHROME_WEBSTORE_CLIENT_SECRET` on your environment and
be part of the "sg chrome ext devs" Google group. (You must also pay Google a one-time fee of $5...)