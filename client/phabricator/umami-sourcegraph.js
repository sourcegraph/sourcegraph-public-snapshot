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

function addAuthorizeButton() {
	var topMenuBar = document.getElementsByClassName("phabricator-main-menu phabricator-main-menu-background");
	if (topMenuBar.length !== 1) {
		return;
	}
	// link container
	var linkContainer = document.createElement("a");
	linkContainer.href = "https://sourcegraph.sgpxy.dev.uberinternal.com/";
	linkContainer.target = "_blank";
	// button
	var authorizeSourcegraphButton = document.createElement("span");
	authorizeSourcegraphButton.innerText = "Authorize Sourcegraph";
	authorizeSourcegraphButton.classList.add("phabricator-wordmark");
	authorizeSourcegraphButton.style.fontSize = "14px";
	authorizeSourcegraphButton.style.marginTop = "13px";
	authorizeSourcegraphButton.style.float = "right";
	authorizeSourcegraphButton.style.marginRight = "40px";
	authorizeSourcegraphButton.style.color = "yellow";
	authorizeSourcegraphButton.title = "Click here to authenticate Sourcegraph and enable code intelligence on Phabricator, then refresh this window. See console log for more details.";
	linkContainer.appendChild(authorizeSourcegraphButton);
	topMenuBar[0].appendChild(linkContainer);
}

function onError(error) {
	console.error("Error loading Sourcegraph Phabricator extension asset from https://sourcegraph.sgpxy.dev.uberinternal.com/.assets/scripts/phabricator.bundle.js.");
	console.error("Please visit https://sourcegraph.sgpxy.dev.uberinternal.com/, authenticate and log-in, and accept the certificate to enable code intelligence.");
	console.error("Then, refresh this window.");
	addAuthorizeButton();
}

var phabricatorUsername = getPhabricatorUsername();
if (phabricatorUsername && pilotUsers[phabricatorUsername]) {
	var script = document.createElement("script");
	script.type = "text/javascript";
	script.defer = true;
	// this url should point to the sourcegraph instance serving the phabricator tooltips
	// eventually, this should be umami.bundle.js, but the first version of the script we 
	// shipped was phabricator.bundle.js
	script.src = "https://sourcegraph.sgpxy.dev.uberinternal.com/.assets/scripts/phabricator.bundle.js";
	script.onerror = onError;
	document.getElementsByTagName("head")[0].appendChild(script);
}
