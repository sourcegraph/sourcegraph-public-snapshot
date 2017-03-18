/**
* @provides uber-sourcegraph
*/

/** This matches patterns like /p/username */
var USERNAME_URL_PATTERN = /\/p\/([A-Z0-9-_]+)/i;

/**
 * Scrapes a Phabricator username from the DOM.
 */
function getPhabricatorUsername() {
	var coreMenuItems = document.getElementsByClassName("phabricator-core-user-menu");
	for (var i = 0; i < coreMenuItems.length; i++) {
		var coreMenuItem = coreMenuItems.item(i);
		var possiblePersonUrl = coreMenuItem.getAttribute("href");
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

var pilotUsers = {
	/*
	 * Fill in users for the Uber pilot here.
	 */
};

var phabricatorUsername = getPhabricatorUsername();
if (phabricatorUsername && pilotUsers[phabricatorUsername]) {
    var script = document.createElement("script");
    script.type = "text/javascript";
    script.defer = true;
	// this url should point to the sourcegraph instance serving the phabricator tooltips
	// eventually, this should be umami.bundle.js, but the first version of the script we 
	// shipped was phabricator.bundle.js
    script.src = "https://sourcegraph.uberinternal.com/.assets/scripts/phabricator.bundle.js";
    document.getElementsByTagName("head")[0].appendChild(script);
}
