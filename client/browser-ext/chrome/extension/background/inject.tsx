function loadScript(name: string, tabId: number, cb: () => void): Promise<void> {
	if (process.env.NODE_ENV === "production") {
		// Do not use direct tab injection; instead, scripts are loaded via
		// content_scripts (see: https://developer.chrome.com/extensions/content_scripts).
		return Promise.resolve();
	} else {
		// dev: async fetch bundle
		return fetch(`https://localhost:3000/js/${name}.bundle.js`)
			.then((res) => res.text())
			.then((fetchRes) => {
				chrome.tabs.executeScript(tabId, {code: fetchRes, runAt: "document_end"}, cb);
			});
	}
}

const arrowURLs = ["^https://github\\.com", "^https://www\\.github\\.com", "^https://sourcegraph\\.com", "^https://www\\.sourcegraph\\.com"];

if (process.env.NODE_ENV !== "production") {
	chrome.tabs.onUpdated.addListener((tabId: number, changeInfo: chrome.tabs.TabChangeInfo, tab: chrome.tabs.Tab) => {
		if (changeInfo.status !== "loading" || (tab.url && !tab.url.match(arrowURLs.join("|")))) {
			return Promise.resolve();
		}

		// tslint:disable-next-line
		return loadScript("inject", tabId, () => console.log("load inject bundle success!"));
	});
}
