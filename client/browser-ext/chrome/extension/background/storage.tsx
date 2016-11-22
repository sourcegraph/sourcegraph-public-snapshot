// The chrome.cookies and chrome.storage APIs may not be directly accessible
// through by content scripts in Chrome AND Firefox. Instead, content scripts
// may pass message to this background script.
chrome.runtime.onMessage.addListener((message, sender, cb) => {
	if (message.type === "setIdentity") {
		chrome.storage.local.set({identity: message.identity});
	} else if (message.type === "getIdentity") {
		chrome.storage.local.get("identity", (obj) => {
			const {identity} = obj;
			cb(identity);
		});
		return true;
	} else if (message.type === "getSessionToken") {
		chrome.cookies.get({url: "https://sourcegraph.com", name: "sg-session"}, (sessionToken) => {
			cb(sessionToken ? sessionToken.value : null);
		});
		return true;
	}
});
