/// <reference path="../../../globals.d.ts" />

import { sourcegraphUrl } from "../../../app/utils/context";

/**
 * The chrome.cookies and chrome.storage APIs may not be directly accessible
 * through by content scripts in Chrome AND Firefox. Instead, content scripts
 * may pass message to this background script.
 *
 * setIdentity is a message sent from the Sourcegraph.com front end.
 * getIdentity is a message sent from the extension eventLogger
 * getSessionToken gets any logged in token from the sourcegraph.com cookie, so that we can
 * include it with XHR requests, and is sent when first injecting the extension on GitHub.
 */
chrome.runtime.onMessage.addListener((message, _, cb) => {
	if (message.type === "setIdentity") {
		chrome.storage.local.set({ identity: message.identity });
	} else if (message.type === "getIdentity") {
		chrome.storage.local.get("identity", (obj) => {
			const { identity } = obj;
			cb(identity);
		});
		return true;
	} else if (message.type === "getSessionToken") {
		chrome.cookies.get({ url: sourcegraphUrl, name: "sg-session" }, (sessionToken) => {
			cb(sessionToken ? sessionToken.value : null);
		});
		return true;
	}
});
