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
$ yarn
```

## Development

```bash
$ yarn install
$ yarn run dev
```
* Allow `https://localhost:3000` (insecure) connections in Chrome (navigate to https://localhost:3000, click "ADVANCED",
then "Proceed to localhost"). This is necessary because pages are injected on Sourcegraph/GitHub (https), so `webpack-dev-server`
procotol must also be https.
* [Load unpacked extensions](https://developer.chrome.com/extensions/getstarted#unpacked) with `./dev` folder.
* Webpack will manage hot reloading via `react-transform`.

## Deployment to Chrome web store 

1. In `sourcegraph/sourcegraph/client/browser-ext/build/manifest.json`, bump the version on line 2. i.e. if the current version number is `"version": "1.1.98"`, change it to `"version": "1.1.99"`.
2. In `sourcegraph/sourcegraph/client/browser-ext`, run:
```bash
$ make bundle
```
3. Go to https://chrome.google.com/webstore/category/extensions and from settings, go to the developer dashboard: https://cl.ly/1J3c0N1j0E0F. 
4. Click edit on the Sourcegraph for GitHub extension. 
5. Click "Upload Updated Package" and select the newly generated `sourcegraph/sourcegraph/client/browser-ext/chrome-bundle.zip`
6. Publish the Chrome extensions

## Update Phabricator plugin JavaScript

Build the latest phabricator.bundle.js file, and copy to the Sourecgraph assets folder.

```bash
$ make phabricator
```

## Deploy (chrome extension only, ask John/Matt for Firefox)

```bash
$ make deploy
```

To deploy the chrome extension with your Google Apps credentials, you must have `CHROME_WEBSTORE_CLIENT_SECRET` on your environment and
be part of the "sg chrome ext devs" Google group. (You must also pay Google a one-time fee of $5...)

### Steps to deploy
* Remember to bump the version number in manifest.prod.js
* Run browser extension unit tests in the sourcegraph top level directory `test/e2e2`
* Run `make deploy` in `client/browser-ext`

