# Sourcegraph browser extensions

[![build](https://badge.buildkite.com/8aec2156579076111918e632ba9e188cafc8c9598968457f4f.svg)](https://buildkite.com/sourcegraph/browser-extension)
[![chrome version](https://img.shields.io/chrome-web-store/v/dgjhfomjieaadpoljlnidmbgkdffpack.svg)](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack)
[![chrome users](https://img.shields.io/chrome-web-store/users/dgjhfomjieaadpoljlnidmbgkdffpack.svg)](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack)
[![chrome rating](https://img.shields.io/chrome-web-store/rating/dgjhfomjieaadpoljlnidmbgkdffpack.svg)](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack)
[![firefox version](https://img.shields.io/amo/v/sourcegraph.svg)](https://addons.mozilla.org/en-US/firefox/addon/sourcegraph/)
[![firefox users](https://img.shields.io/amo/users/sourcegraph.svg)](https://addons.mozilla.org/en-US/firefox/addon/sourcegraph/)
[![firefox rating](https://img.shields.io/amo/rating/sourcegraph.svg)](https://addons.mozilla.org/en-US/firefox/addon/sourcegraph/)

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
- `npm`
- `make`

## Development

For each browser run:

```bash
$ npm install
$ npm run dev
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
npm run dev:firefox
```

A Firefox window will be spun up with the extension already installed.

#### Updating the bundle

Save a file and wait for webpack to finish rebuilding.

#### Caveats

The window that is spun up is completely separate from any existing sessions you have on Firefox.
You'll have to sign into everything at the begining of each development session(each time you run `npm run dev:firefox`).
You should ensure you're signed into any Sourcegraph instance you point the extension at as well as Github.

### Firefox (manual)

- Go to `about:debugging`
- Select "Enable add-on debugging"
- Click "Load Temporary Add-on" and select "firefox-bundle.xpi"
- [More information](https://developer.mozilla.org/en-US/docs/Tools/about:debugging#Add-ons)

#### Updating the bundle

Click reload for Sourcegraph at `about:debugging`

### Safari

- Make sure developer tools are enabled
  - Open Safari > Preferences (or `Cmd+,`) and click Advanced
  - Check "Show Develop menu in menu bar"
- Open `Develop -> Show Extension Builder`
- Click the `+` at the bottom left of the Extension Builder and select `browser-extension/Sourcegraph.safariextension`
- Click `Install` and `Add extension...` and open `~/path/to/browser-extension/Sourcegraph.safariextension`
- To test against localhost, follow https://about.sourcegraph.com/docs/features/safari-extension:
  - Set `"corsOrigin": "https://github.com"` in your localhost site configuration
  - Download the `ngrok` executable https://dashboard.ngrok.com/get-started, authenticate, and run `./ngrok localhost 3080`
  - Visit the URL it spits out, sign in, and add the URL to your Safari extension (to the left of the address bar - don't change Settings in the Extension Builder)

## Testing

Coming soon...

## Deploy (Chrome)

- Ensure that you have bumped and committed the version in `chrome/extension.info.json`
- Run `npm run build` to generate the new production build.
- Sign in via Google with your Sourcegraph email address.
- Naviate to https://chrome.google.com/webstore/developer/dashboard?pli=1
- If you don't see the extension, ask [to get added to the `sg-chrome-ext-devs` Google Group](https://groups.google.com/forum/#!managemembers/sg-chrome-ext-devs/invite)
- Click "edit" in the row associated with Sourcegraph for GitHub
- Click "Upload Updated Package" in the top section (inside Upload).
- Click choose file and select the `build/bundles/chrome-bundle.zip` file.
- Add release notes and submit the build. It will be availble for the submitter instantly, but users will see it in a couple of hours. If this is a big fix for a bug then it is worth telling users they can go to `chrome://extensions` and clicking "Update extensions now" (This option is only available if "Developer mode" is enabled).
- Click "Publish Changes"

### Automated

To deploy the chrome extension with your Google Apps credentials, you must have `CHROME_WEBSTORE_CLIENT_SECRET` on your environment and
be part of the "sg chrome ext devs" Google group. (You must also pay Google a one-time fee of $5...)

```bash
$ make deploy
```

## Deploy (Safari)

- Ensure you have the production build by running `npm run build`
- Open the extension builder in Safari `Develop -> Show Extension Builder`
  - Notice the `Sourcegraph.safariextz` zip file is created in `browser-extension`
- Click `Build Package`
- Open the [Safari extension bucket](https://console.cloud.google.com/storage/browser/sourcegraph-safariextz?project=sourcegraph-dev) on GCP
- Click `UPLOAD FILES` and pick `Sourcegraph.safariextz`
- In the `Resolve Conflict` prompt, choose to replace the existing object
- `IMPORTANT` - Ensure that the `Share Publicly` setting is still checked
- Click the `...` menu on the far right of the line item for `Sourcegraph.safariextz`
- Copy `Sourcegraph.safariextz` and change the destination name to `Sourcegraph-<version>.safariextz`
  - Make sure to keep the source permissions

## Deploy (Firefox)

- Sign into Firefox under a Sourcegraph developer account
- Go to https://addons.mozilla.org/en-US/developers/addon/sourcegraph/versions/submit/
- Ensure you have completed the steps above for release a production version. The three most important steps are: 1) Ensure it runs and there are no errors 2) Bump the version. 3) The `dist` build reflects the current changes you've made.
- Click "Select a file..." and click the `build/bundles/firefox-bundle.xpi` that was generated from `npm run build`.
- Following the upload, delete all contents of the `node_modules` directory except the `@sourcegraph` directory.
  - Firefox requires us to upload our source code because the files are bundled. They also do not allow you to upload the `node_modules` directory unless the extension relies on private modules. We use private node modules so we need to include the `node_modules/@sourcegraph` contents to ensure the extension is approved.
- Create a zip of the entire browser-extension repository.
- Upload the browser-extension.zip to the Firefox store before clicking continue.
- Append to the version notes if there is something significant. Otherwise use the previous version notes. You can find previous version notes by going to https://addons.mozilla.org/en-US/developers/addon/sourcegraph/versions and clicking the previous submission.
- ALWAYS INCLUDE NOTES FOR REVIEWERS: (Copy paste what is below, if a significant change happened include it and update the README.)
  Running from source:

1.  With npm 6 installed, run `npm install`
2.  npm run build
3.  Go to about:debugging
4.  Select "Enable add-on debugging"
5.  Load Temporary Add-on
6.  Click "Load Temporary Add-on" and select "build/bundles/firefox-bundle.xpi"

NOTE: This extension includes an opt-in for event tracking on GitHub.com for the purposes of personalization.

How to use the extension:
The Sourcegraph developer extension works on GitHub.com. Below you will find a list of Sourcegraph's features with relevant URLs as well as screenshots and videos to help verify the extension.

Relevant URLs:
Open on Sourcegraph URL: https://github.com/gorilla/mux
Code browsing URL: https://github.com/gorilla/mux/blob/master/mux.go

Main features of Sourcegraph:

1.  Open in Sourcegraph - This feature takes the user from the repository homepage to viewing the repository on Sourcegraph.com https://goo.gl/jepnDz
2.  Code intelligence when hovering over code tokens - This provides developers with IDE like code browsing tools on GitHub - https://goo.gl/G1cMMM
3.  Action items for hover tooltip - Users can now see references, the definition, and also search based on the currently selected token. -
    https://goo.gl/CHFnjr
4.  File tree navigation when viewing a GitHub repository - https://goo.gl/7NafYf

- Click Save Changes
- You're done.
