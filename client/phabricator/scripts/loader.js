/**
 * Scrapes a Phabricator username from the DOM.
 */
function getPhabricatorUsername() {
  var USERNAME_URL_PATTERN = /\/p\/([A-Z0-9-_]+)/i;
  var coreMenuItems = document.getElementsByClassName('phabricator-core-user-menu');
  for (var i = 0; i < coreMenuItems.length; i++) {
    var coreMenuItem = coreMenuItems.item(i);
    var possiblePersonUrl = coreMenuItem.getAttribute('href');
    if (!possiblePersonUrl) {
      continue;
    }
    var match = USERNAME_URL_PATTERN.exec(possiblePersonUrl);
    if (!match) {
      continue;
    }
    return match[1];
  }
  return null;
}

/**
 * To prevent loading the extension for all users, specifify a user whitelist
 * by changing this line to `var userWhitelist = { "username": true, ... };`
 */
var userWhitelist = undefined;

/**
 * To load the extension from a different endpoint than the one set during
 * installation, add domains to this whitelist. Don't remove `window.SOURCEGRAPH_URL`
 * as this is the default URL set by the Phabricator admin during installation.
 * e.g. `var sourcegraphWhitelist = [window.SOURCEGRAPH_URL, 'https://test.sourcegagraph.mycompany.org']`
 */
var sourcegraphWhitelist = [window.SOURCEGRAPH_URL];

/**
 * Installing the loader script requires specifying the Sourcegraph Server URL.
 * This is the default endpoint used by the loader and for API requests from the extension.
 * Override the default by setting `window.localStorage.SOURCEGRAPH_URL = ...` (requires adding a value
 * to `sourcegraphWhitelist` above).
 */
var sourcegraphURL = window.localStorage.SOURCEGRAPH_URL || window.SOURCEGRAPH_URL;
if (!sourcegraphWhitelist.includes(sourcegraphURL)) {
  // The URL is not included in the whitelist, so fail.
  throw new Error('cannot load Sourcegraph extension for Phabricator from ' + sourcegraphURL);
}

window.SOURCEGRAPH_URL = sourcegraphURL; // possibly override value set by installation

function load() {
  var script = document.createElement('script');
  script.type = 'text/javascript';
  script.defer = true;
  script.src = sourcegraphURL + '/.assets/extension/scripts/phabricator.bundle.js';
  document.getElementsByTagName('head')[0].appendChild(script);

  var head = document.head || document.getElementsByTagName('head')[0];
  var styleLink = document.createElement('link');
  styleLink.rel = 'stylesheet';
  styleLink.href = sourcegraphURL + '/.assets/extension/css/style.bundle.css';
  head.appendChild(styleLink);
}

if (userWhitelist) {
  // Load the extension iff the current user is on the whitelist.
  var username = getPhabricatorUsername();
  if (username && userWhitelist[username]) {
    load();
  }
} else {
  // Unconditionally load the extension.
  load();
}
