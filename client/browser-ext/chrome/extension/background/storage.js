// The chrome.storage API is directly accessible by content scripts
// in chrome, but not firefox. As a workaround, content scripts
// read/write to chrome.storage by passing a message to this background
// script.
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
	}
});
