// The chrome.storage API is directly accessible by content scripts
// in chrome, but not firefox. As a workaround, content scripts
// read/write to chrome.storage by passing a message to this background
// script. Also handles chrome.cookies for session and csrf tokens
// from sourcegraph.com
chrome.runtime.onMessage.addListener((message, sender, cb) => {
	if (message.type === "get") {
		chrome.storage.local.get("state", (obj) => {
			const {state} = obj;
			const initialState = JSON.parse(state || "{}");
			cb(initialState);
		});
		return true; // signal asynchronous response
	} else if (message.type === "set") {
		chrome.storage.local.set({state: message.state});
	} else if (message.type === "setIdentity") {
		chrome.storage.local.set({identity: message.identity});
	} else if (message.type === "getIdentity") {
		chrome.storage.local.get("identity", (obj) => {
			const {identity} = obj;
			cb(identity);
		});
		return true;
	} else if (message.type === "getSessionToken") {
		chrome.cookies.get({url: "https://sourcegraph.com", name: "sg-session"}, (sessionToken) => {
			cb({sessionToken: sessionToken ? sessionToken.value : null});
		});
		return true;
	}
});
