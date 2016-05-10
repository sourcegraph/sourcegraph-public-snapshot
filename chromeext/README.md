# Sourcegraph browser extensions for Google Chrome and Firefox

## Project structure

```
server
├── app <-- React/Redux applicaiton providing definition-based code search
├── chrome
│   └── extension
│       └── background
│       	└── inject.js <-- helper functions for development only
│       	└── storage.js <-- wrapper for chrome.storage get/set
│       └── annotations.js <-- logic for linkifying GitHub code with Sourcegraph annotation data
│       └── background.js <-- background script for development workflows & storage
│       └── inject.js <-- manages app/annotation injection
│   └── views <- extension "pages" (for development only)
│   └── manifest.dev.json <-- supports hot-reloading code (requires extra permissions)
│   └── manifest.prod.json <-- manifest for production build, requires minimal permissions
├── scripts <-- build/development scripts
├── webpack <-- build configuration
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

## Build

```bash
$ npm run build
```

#### Firefox

To bundle the extension for Firefox, compress the *contents* of the `/build` folder (not the folder itself). Rename the file `sourcegraph.xpi`.

## Compress (chrome only)

```bash
$ npm run build
$ npm run compress -- [options]
```

#### Options

If you want to build a `crx` file, provide options and add an `update.xml` file url in [manifest.json](https://developer.chrome.com/extensions/autoupdate#update_url manifest.json).

* --app-id: the extension id
* --key: private key path (default: './key.pem')
  you can use `npm run compress-keygen` to generate private key `./key.pem`
* --codebase: the `crx` file url

See [autoupdate guide](https://developer.chrome.com/extensions/autoupdate) for more information.

## Boilerplate

This project was adapted from this boilerplate: https://github.com/jhen0409/react-chrome-extension-boilerplate.
