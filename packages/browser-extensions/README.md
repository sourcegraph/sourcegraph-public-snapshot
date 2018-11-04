# Sourcegraph browser extensions

[![build](https://travis-ci.org/sourcegraph/browser-extensions.svg?branch=master)](https://travis-ci.org/sourcegraph/browser-extensions)
[![dependencies](https://david-dm.org/sourcegraph/browser-extensions/status.svg)](https://david-dm.org/sourcegraph/browser-extensions)
[![code style: prettier](https://img.shields.io/badge/code_style-prettier-ff69b4.svg)](https://github.com/prettier/prettier)
[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release)
![license](https://img.shields.io/badge/license-MIT-blue.svg)

[![chrome version](https://img.shields.io/chrome-web-store/v/dgjhfomjieaadpoljlnidmbgkdffpack.svg?logo=Google%20Chrome&logoColor=white)](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack)
[![chrome users](https://img.shields.io/chrome-web-store/users/dgjhfomjieaadpoljlnidmbgkdffpack.svg)](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack)
[![chrome rating](https://img.shields.io/chrome-web-store/rating/dgjhfomjieaadpoljlnidmbgkdffpack.svg)](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack)\
[![firefox version](https://img.shields.io/amo/v/sourcegraph.svg?logo=Mozilla%20Firefox&logoColor=white)](https://addons.mozilla.org/en-US/firefox/addon/sourcegraph/)
[![firefox users](https://img.shields.io/amo/users/sourcegraph.svg)](https://addons.mozilla.org/en-US/firefox/addon/sourcegraph/)
[![firefox rating](https://img.shields.io/amo/rating/sourcegraph.svg)](https://addons.mozilla.org/en-US/firefox/addon/sourcegraph/)

## Overview

The Sourcegraph browser extension adds tooltips to code on GitHub, Phabricator, and Bitbucket.
The tooltips include features like:

- symbol type information & documentation
- go to definition & find references (currently for Go, Java, TypeScript, JavaScript, Python)
- find references

#### 🚀 Install: [**Sourcegraph for Chrome**](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack) — [**Sourcegraph for Firefox**](https://addons.mozilla.org/en-US/firefox/addon/sourcegraph/)

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

- `src/extension/`
  - Entrypoint for browser extension builds. (Includes bundled assets, background scripts, options)
- `src/browser`
  - [A wrapper around the browser APIs.](./src/browser/README.md)
- `src/libs/`
  - Isolated pieces of the browser extension. This contains code that is specific to code hosts and separate "mini applications" included in the browser extension such as the `src` omnibar cli.
- `src/libs/phabricator/`
  - Entrypoint for Phabricator extension. This is used by the browser extension and [sourcegraph/phabricator-extension](https://github.com/sourcegraph/phabricator-extension).
- `src/shared/`
  - Code shared by the extension and the libraries. Ideally, nothing in here should reach into any other directory.
- `src/config/`
  - Polyfills and configuration/plumbing code that is bundled via webpack. The configuration code adds properties to `window` that make it easier to tell what environment the script is running in. This is useful because the code can be run in the content script, background, options page, or in the actual page when injected by Phabricator and each environment will have different ways to do different things.
- `cypress`
  - E2e test suite.
- `scripts/`
  - Development scripts.
- `webpack`
  - Build configs.
- `build`
  - Generated directory containing the output from webpack and the generated bundles for each browser.

## Requirements

- `node`
- `yarn`
- `make`

## Development

For each browser run:

```bash
$ yarn
$ yarn run dev
```

Now, follow the steps below for the browser you intend to work with.

### Chrome

- Browse to [chrome://extensions](chrome://extensions).
- If you already have the Sourcegraph extension installed, disable it by unchecking the "Enabled" box.
- Click on [Load unpacked extensions](https://developer.chrome.com/extensions/getstarted#unpacked), and select the `build/chrome` folder.
- Browse to any public repository on GitHub to confirm it is working.
- After making changes it is necessary to refresh the extension. This is done by going to [chrome://extensions](chrome://extensions) and clicking "Reload".

![Add dist folder](readme-load-extension-asset.png)

#### Updating the bundle

Click reload for Sourcegraph at `chrome://extensions`

### Firefox (hot reloading)

In a separate terminal session run:

```bash
yarn run dev:firefox
```

A Firefox window will be spun up with the extension already installed.

#### Updating the bundle

Save a file and wait for webpack to finish rebuilding.

#### Caveats

The window that is spun up is completely separate from any existing sessions you have on Firefox.
You'll have to sign into everything at the begining of each development session(each time you run `yarn run dev:firefox`).
You should ensure you're signed into any Sourcegraph instance you point the extension at as well as Github.

### Firefox (manual)

- Go to `about:debugging`
- Select "Enable add-on debugging"
- Click "Load Temporary Add-on" and select "firefox-bundle.xpi"
- [More information](https://developer.mozilla.org/en-US/docs/Tools/about:debugging#Add-ons)

#### Updating the bundle

Click reload for Sourcegraph at `about:debugging`

## Testing

Coming soon...

## Deploy

Deployment to Firefox and Chrome extension stores happen automatically in CI when the `release` branch is updated.
Releases are also uploaded to the [GitHub releases page](https://github.com/sourcegraph/browser-extensions/releases) and tagged in git.

Make sure that commit messages follow the [Conventional Commits Standard](https://conventionalcommits.org/) as the commit message will be used for the (public) release notes and to automatically determine the version number.

To release the latest commit on master, ensure your master is up-to-date and run

```sh
git push origin master:release
```
