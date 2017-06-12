# Sourcegraph browser extensions for Google Chrome and Firefox

## Overview

The Sourcegraph browser extension adds tooltips to code on GitHub, Phabricator, and Bitbucket.
The tooltips include features like:
  - symbol type information & documentation
  - go to definition & find references (currently for Go, Java, TypeScript, JavaScript, Python)
  - find references
  - improved search functionality

It works as follows:
  - when visiting e.g. https://github.com/..., the extension injects a content script (inject.bundle.js)
  - there is a background script running to access certain chrome APIs, like storage (background.bundle.js)
  - a "code view" contains rendered (syntax highlighted) code (in an HTML table); the extension adds event listeners to the code view which control the tooltip
  - when the user mouses over a code table cell, the extension modifies the DOM node:
    - text nodes are wrapped in <span> (so hover/click events have appropriate specificity)
    - element nodes may be recursively split into multiple element nodes (e.g. a <span>&Router{namedRoutes:<span> contains multiple code tokens, and event targets need more granular ranges)
    - ^^ we assume syntax highlighting takes care of the base case of wrapping a discrete language symbol
    - tooltip data is fetched from the Sourcegraph API
  - when an event occurs, we modify a central state store about what kind of tooltip to display
  - code subscribes to the central store updates, and creates/adds/removes/hides an absolutely positioned element (the tooltip)

## Project Layout

- `app/`
  - application code, e.g. injected onto GitHub (as a content script)
- `chrome/`
  - entrypoint for Chrome extension. Includes bundled assets, background scripts, options)
- `phabricator/`
  - entrypoint for Phabricator extension. The Phabricator extension is injected by Phabricator (not Chrome)
- `scripts/`
  - development scripts
- `test/`
  - test code
- `webpack`
  - build configs

## Requirements

- `node`
- `yarn`
- `make`

## Development (Chrome, with hot reloading)

```bash
$ yarn install
$ yarn run dev
```

* Allow `https://localhost:3000` (insecure) connections in Chrome (navigate to https://localhost:3000, click "ADVANCED",
then "Proceed to localhost"). This is necessary because pages are injected on Sourcegraph/GitHub (https), so `webpack-dev-server`
procotol must also be https.
* [Load unpacked extensions](https://developer.chrome.com/extensions/getstarted#unpacked) with `./dev` folder.

## Development (Firefox)

* `make bundle`
* Go to `about:debugging`
* Select "Enable add-on debugging"
* Click "Load Temporary Add-on" and select "firefox-bundle.xpi"
* [More information](https://developer.mozilla.org/en-US/docs/Tools/about:debugging#Add-ons)

## Testing

Coming soon...

## Deploy (Chrome)

### Manual

1. In `sourcegraph/sourcegraph/client/browser-ext/build/manifest.json`, bump the version on line 2. i.e. if the current version number is `"version": "1.1.98"`, change it to `"version": "1.1.99"`.
1. In `sourcegraph/sourcegraph/client/browser-ext`, run `make bundle`
1. Go to https://chrome.google.com/webstore/category/extensions and from settings, go to the developer dashboard: https://cl.ly/1J3c0N1j0E0F.
1. Click edit on the Sourcegraph for GitHub extension.
1. Click "Upload Updated Package" and select the newly generated `sourcegraph/sourcegraph/client/browser-ext/chrome-bundle.zip`
1. Publish (at the bottom of the form)

### Automated

To deploy the chrome extension with your Google Apps credentials, you must have `CHROME_WEBSTORE_CLIENT_SECRET` on your environment and
be part of the "sg chrome ext devs" Google group. (You must also pay Google a one-time fee of $5...)

```bash
$ make deploy
```

## Deploy (Firefox)

Coming soon...
