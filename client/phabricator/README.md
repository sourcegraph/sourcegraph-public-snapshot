# Sourcegraph extension for Phabricator

## Overview

This extension provides tooltips on Phabricator like the tooltips provided by our
[Chrome extension for GitHub](https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en).

Tooltips are provided on:
- diffusion views like https://phabricator.mycompany.com/source/nrepo/browse/master/file.go and https://phabricator.mycompany.com/source/nrepo/change/master/file.go
- differential pages like https://phabricator.mycompany.com/D1
- commit views like https://phabricator.mycompany.com/rNREPOf9a34fd4ddd26824e4653d1f473c5709f2c21bd8

### How does it work?

Whenever you load a Phabricator page, the extension will load additional JavaScript & CSS assets. These
assets contain the logic and styles for marking up Phabricator views w/ Sourcegraph tooltips.

Data for the tooltips is fetched via cross-origin XHR requests to your Sourcegraph Server. Requests are made with
the user's Sourcegraph cookie, so users must first sign in to Sourcegraph to receive tooltips while on Phabricator.

## Requirements

- A Phabricator installation
- A Sourcegraph Server installation

## Installation

First, copy the contents of the `scripts` directory to your Phabricator server.

### Loading assets without automatic updates

For most users, a manual installation & upgrade workflow is preferred. Follow these steps
(which may need to be run with `sudo`):

* Update `/var/www/phabricator/src/applications/base/controller/PhabricatorController.php` to include the following lines at the top of `willBeginExecution()` (this instructs Phabricator to load Sourcegraph's JavaScript and CSS assets):
```
require_celerity_resource("sourcegraph");
require_celerity_resource("sourcegraph-style");
```
* Run `./install_bundle.sh $SOURCEGRAPH_SERVER_URL` (this adds the JavaScript and CSS assets to Phabricator's static asset map)

### Loading assets with automatic updates

You may configure the extension to load assets dynamically from your Sourcegraph Server, in which case updating
your server can automatically update the JavaScript and CSS which gets loaded on Phabricator pages.

* Update `/var/www/phabricator/src/applications/base/controller/PhabricatorController.php` to include the following lines at the top of `willBeginExecution()` (this instructs Phabricator to load Sourcegraph's JavaScript shim):
```
require_celerity_resource("sourcegraph");
```
* Run `./install_loader.sh $SOURCEGRAPH_SERVER_URL` (this adds a JavaScript shim to Phabricator's static asset map which loads assets from your Sourcegraph Server)


## Development

The Phabricator extension shares code with our Chrome extension for GitHub, so one way to develop it is as a Chrome extension.
If this would interfere with an installed version of the extension, you can set a flag which prevents running the installed
extension assets (since you want the Chrome extension to do that instead):

```
window.localStorage['SOURCEGRAPH_DISABLED'] = true;
```

Alternatively, you can develop the extension using the loader shim and setting your Sourcegraph Server to localhost, like this:

```
./install_loader.sh http://localhost:3080
```

Then run a build loop for the Phabricator extension which outputs generated files to the Sourcegraph frontend's static assets root.
