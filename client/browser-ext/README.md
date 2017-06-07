# Sourcegraph browser extensions for Google Chrome and Firefox

## Overview

The Sourcegraph browser extension marks up "code view" on GitHub, Phabricator, and
Bitbucket. Here's how it works at a high level:

- when visiting e.g. https://github.com/..., the browser extension "injects" a few React components on the page
- "code views" refer to pages which show code; it seems the standard way to show code (for diffs or otherwise) is using an HTML table
- within the HTML table, each data cell may contain e.g. a line number, empty space, a "context expander", or actual code
- the "BlobAnnotator" component determines information about the code view (the path, the revisision, etc) and registers event listeners on the table data cells which contain code
- the first time the user hovers over a code data cell (where a listener has been registered), the extension modifies the code cell according to this logic:
  - text nodes are wrapped in <span> (this is necessary because a hover/click event can't have a text node target, and the extension needs to know exactly where the user hovered or clicked)
  - element nodes may be split into multiple element nodes, according to parsing logic (in some cases, there may be something like a <span>&Router{namedRoutes:<span> which actually contains multiple distinct code tokens -- but a hover/click handler won't know pricely which token was references unless they are split)
  - recurse on element nodes with children (e.g. <span><span>&Router{namedRoutes:</span><span>: make(map[string]*Route)}</span>)
- (by lazily modifying code cells only when the user hovers over them, page performance doesn't suffer)
- we use RxJS to elegantly handle streams of mousein/out/over/up/click event handlers, async tooltip fetching logic, etc and thereby determine "what should the tooltip be right now"
- the "store.tsx" module is a poor-mans Redux implementation, also using RxJS; the "tooltip.tsx" module subscribes to the store and displays an appropriate tooltip for the current state of the store

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

