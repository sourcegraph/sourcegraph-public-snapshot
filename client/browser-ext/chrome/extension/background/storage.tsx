import { sourcegraphUrl } from "../../../app/utils/context";

const application = "com.sourcegraph.browser_ext_host";
let port = null;

port = chrome.runtime.connectNative(application);

port.onMessage.addListener((e) => console.log("port connected", e));

port.onDisconnect.addListener((e) => {
	console.log('unexpected disconnect', e);
	port = null;
});

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
	switch (message.type) {
		case "setIdentity":
			chrome.storage.local.set({ identity: message.identity });
			return;

		case "getIdentity":
			chrome.storage.local.get("identity", (obj) => {
				const { identity } = obj;
				cb(identity);
			});
			return true;

		case "getSessionToken":
			chrome.cookies.get({ url: sourcegraphUrl, name: "sg-session" }, (sessionToken) => {
				cb(sessionToken ? sessionToken.value : null);
			});
			return true;

		case "openSourcegraphTab":
			chrome.tabs.query({ url: "https://sourcegraph.com/*" }, (tabs) => {
				if (tabs.length > 0) {
					const tab = tabs[0];
					chrome.tabs.update(tab.id, { active: true }, () => {
						chrome.tabs.executeScript(tab.id, { code: `window.dispatchEvent(new CustomEvent("browser-ext-navigate", {detail: {url: "${message.url}"}}))` });
					});
					cb(true);
				} else {
					cb(false);
				}
			});
			return true;

		case "openEditor":
			const msg = { cmd: message.cmd };
			if (port) {
				port.postMessage(msg);
			} else {
				chrome.runtime.sendNativeMessage(application, msg, cb);
			}
			return true;
	}
});
