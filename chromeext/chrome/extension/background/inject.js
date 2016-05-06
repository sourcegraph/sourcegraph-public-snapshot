function isInjected(tabId) {
	return chrome.tabs.executeScriptAsync(tabId, {
		code: `var injected = window.reactExampleInjected;
			window.reactExampleInjected = true;
			injected;`,
		runAt: "document_start"
	});
}

function loadScript(name, tabId, cb) {
	if (process.env.NODE_ENV === "production") {
		// Do not use direct tab injection; instead, scripts are loaded via
		// content_scripts (see: https://developer.chrome.com/extensions/content_scripts).
	} else {
		// dev: async fetch bundle
		fetch(`https://localhost:3000/js/${name}.bundle.js`)
			.then((res) => res.text())
			.then((fetchRes) => {
				// Load redux-devtools-extension inject bundle,
				// because inject script and page is in a different context
				const request = new XMLHttpRequest();
				request.open("GET", "chrome-extension://lmhkpmbekcpmknklioeibfkpmmfibljd/js/inject.bundle.js");  // sync
				request.send();
				request.onload = () => {
					if (request.readyState === XMLHttpRequest.DONE && request.status === 200) {
						chrome.tabs.executeScript(tabId, {code: request.responseText, runAt: "document_start"});
					}
				};
				chrome.tabs.executeScript(tabId, {code: fetchRes, runAt: "document_end"}, cb);
			});
	}
}

const arrowURLs = ["^https://github\\.com", "^https://www\\.github\\.com", "^https://sourcegraph\\.com", "^https://www\\.sourcegraph\\.com"];

if (process.env.NODE_ENV !== "production") {
	chrome.tabs.onUpdated.addListener(async function(tabId, changeInfo, tab) {
		if (changeInfo.status !== "loading" || !tab.url.match(arrowURLs.join("|"))) return;

		const result = await isInjected(tabId);
		if (chrome.runtime.lastError || result[0]) return;

		loadScript("inject", tabId, () => console.log("load inject bundle success!"));
	});
}
