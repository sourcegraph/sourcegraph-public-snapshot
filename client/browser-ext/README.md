# Sourcegraph browser extensions for Google Chrome and Firefox

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

Generate zipped artifacts (`chrome-bundle.zip` and `firefox-bundle.xpi`) to upload to the Chrome/Firefox web stores.

```bash
$ make bundle
```

## Deploy (chrome extension only)

```bash
$ make deploy
```

To deploy the chrome extension with your Google Apps credentials, you must have `CHROME_WEBSTORE_CLIENT_SECRET` on your environment and
be part of the "sg chrome ext devs" Google group. (You must also pay Google a one-time fee of $5...)