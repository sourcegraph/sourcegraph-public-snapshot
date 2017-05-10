/**
 * Helpers
 */
function getSourcegraphURLInput(): HTMLInputElement {
	return document.getElementById("sourcegraph_url") as HTMLInputElement;
}

/**
 * Initialization
 */
chrome.storage.sync.get((items) => {
	if (!items.sourcegraphURL) {
		chrome.storage.sync.set({ sourcegraphURL: "https://sourcegraph.com" });
	} else {
		getSourcegraphURLInput().value = items.sourcegraphURL;
	}
});
getSourcegraphURLInput().focus();

/**
 * Sync storage value to UI
 */
chrome.storage.onChanged.addListener(function (changes, namespace) {
	getSourcegraphURLInput().value = changes.sourcegraphURL.newValue;
});

/**
 * UI listeners
 */

getSourcegraphURLInput().addEventListener("input", (evt) => {
	chrome.storage.sync.set({
		sourcegraphURL: (evt.target as HTMLInputElement).value
	});
});

getSourcegraphURLInput().addEventListener("keydown", (evt) => {
	if (evt.keyCode === 13) {
		evt.preventDefault();
	}
});
