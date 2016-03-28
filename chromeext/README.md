# Sourcegraph Chrome Extension

This extension enhances file pages on GitHub by annotating code with links to usage examples and documentation. It also adds a keyboard shortcut (Shift+T) that allows you to search for functions, classes, and other code definitions in the repository (in lieu of the text-based search on GitHub).  

You can also browse code on Sourcegraph itself at https://sourcegraph.com.

Currently, Sourcegraph supports <b>Go</b> and <b>Java</b>.  This extension will work on all public Go and Java code as well as most of the popular repositories in Python and JavaScript.  Support for more languages will be rolled out soon - stay tuned!   

To use code search on private repositories, code must first be mirrored onto Sourcegraph.com (for free). 

SECURITY MATTERS: This extension never sends any information about private
repositories to Sourcegraph.

## Development

## Prerequisites

Install [Gulp](https://github.com/gulpjs/gulp/blob/master/docs/getting-started.md#getting-started) and [NodeJS](https://nodejs.org/en/download/), then install the dependencies:
```
cd chrome/
npm install
```

## Building

To build and watch the source files for changes, run `gulp`. Then go to `chrome://extensions`
in Chrome, check "Developer mode" and use *Load Unpacked Extension* to load the
`chrome/build` extension directory.

To reload the Chrome extension when you change files, install
[Extensions Reloader](https://chrome.google.com/webstore/detail/fimgfedafeadlieiabdeeaodndnlbhid)
and run `gulp watch`.

To inject content from http://localhost:3080 instead of from
https://sourcegraph.com, run `DEV=1 grunt`. GitHub is HTTPS-only, so you may
need to [allow mixed content](http://superuser.com/a/487772) temporarily.

## Publishing

To publish a new version:

1. Visit the [Chrome Web Store](https://chrome.google.com/webstore).
1. Click the Settings gear in top right corner -> `Developer Dashboard`.
1. Choose the right extension with the `Status` of `Published` and click `Edit`.
1. Bump the version in `manifest.json` to be greater than the current one.
1. Run `gulp build` to build the extension -> test that it works by loading it into Chrome via *Load Unpacked Extension*!
1. `cd build/ && zip -r ../../release.zip .` to create the zip.
1. Click `Upload Updated Package` on the developer dashboard to upload the new `release.zip`.
1. Scroll to bottom of page and hit `Publish Changes`.



