# Sourcegraph browser extensions for Google Chrome and Firefox

## Project structure

```
browser-ext
├── app <-- [React](https://facebook.github.io/react/) + [Redux](http://redux.js.org/) application
│	└── actions <-- methods to fetch from sourcegraph API & change state
│	└── analytics <-- user event logger
│	└── components <-- "receive" state from reducers as and re-render whenever a property it subscribes to
│						are updated
│	└── constants <-- the names given to actions
│	└── reducers <-- "holder" of current state: functions change current state when actions are dispatched
│	└── store  <-- "persistence" of current state: browser local storage automatically syncs reducer state;
│					extries can be given TTL, though currently this a bit of boilerplate
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

The script (injects.js) will inject application components (better better
thought of as "modules") which may either have UI (as in normal React) or
instead only have side effects (e.g. the "BlobAnnotator" doesn't render
itself, but it is responsible for manipulating the blob dom to include
identifier links).

In both cases, we do this with React.Component as the "modules" base and
rely on its normal lifecycle methods to initiate most of the control flow.
The state container is vanilla Redux, and any component/module can subscribe
to any property on the reducer state and go through a "re-render" cycle when that
property changes. Of course, the component will also re-render if it changes
its internal state through this.setState() or if its parent Component updates
a property, just as in normal React.

In addition to the standard Redux reducer (in-memory) state, we synchronize some
state to browser local storage. This allows for more seamless coordination between
multiple tabs. Currently, only the user access token is synchronized.

Actions are provided on a Component's this.props via the @connect decorator.
Use these to make API requests and update application state.

## Requirements

- `npm` >= 3.6.0
- `node` >= 5.6.0

The latest stable version of node will suffice, which you can install as follows after you have installed npm:

```
sudo npm cache clean -f
sudo npm install -g n
sudo n stable

sudo ln -sf /usr/local/n/versions/node/<VERSION>/bin/node /usr/bin/node
```

## Installation

```bash
$ npm install
```

## Development

```bash
$ npm run dev
```
* Allow `https://localhost:3000` (insecure) connections in Chrome (navigate to https://localhost:3000, click "ADVANCED", then "Proceed to localhost"). This is necessary because pages are injected on Sourcegraph/GitHub (https), so `webpack-dev-server` procotol must also be https.
* [Load unpacked extensions](https://developer.chrome.com/extensions/getstarted#unpacked) with `./dev` folder.
* Webpack will manage hot reloading via `react-transform`.

#### Using Redux DevTools Extension

You can use [redux-devtools-extension](https://github.com/zalmoxisus/redux-devtools-extension) in development mode.
You can also uncomment the code in configureStore.dev.js to have action/state change logging in the
dev console, though it can get a little verbose.

## Build

```bash
$ npm run build # create unzipped distribution artifact in ./build/; required for e2e tests
```

## Test

```bash
$ npm run test # unit tests
$ npm run test-e2e # end-to-end tests; requires running the build step prior
```

## Create distributions

```bash
$ npm run dist
```

The reason this process does more than just compress the build directory is to ensure that the environment that the dist is created mimics the environment for the extension submission teams. It spins up a Docker Linux image, chooses a specific version of node and npm, and runs the build process before compressing the file.
