const telligent = require("telligent-tracker");
const telligentFunctionName = "telligent";

// Create the initializing function
window[telligentFunctionName] = function() {
	(window[telligentFunctionName].q = window[telligentFunctionName].q || []).push(arguments);
};

// Set up the initial queue, if it doesn't already exist
window[telligentFunctionName].q = new telligent.Telligent((window[telligentFunctionName].q || []), telligentFunctionName);

// Must be called once upon initialization
window.telligent("newTracker", "SourcegraphExtensionTracker", "sourcegraph-logging.telligentdata.com", {
	encodeBase64: false,
	appId: "SourcegraphExtension",
	platform: "BrowserExtension",
	env: process.env.NODE_ENV,
	forceSecureTracker: true,
});

window.telligent("addStaticMetadata", "installed_chrome_extension", "true", "userInfo");

chrome.runtime.onMessage.addListener(
	function(request, sender, sendResponse) {
		if (request.type === "trackEvent") {
			window.telligent("track", request.payload.eventAction, request.payload);
		} else if (request.type === "setTrackerUserId") {
			window.telligent("setUserId", request.payload);
		} else if (request.type === "setTrackerDeviceId") {
			window.telligent("addStaticMetadataObject", {deviceInfo: {TelligentWebDeviceId: request.payload}});
		} else if (request.type === "setTrackerGAClientId") {
			window.telligent("addStaticMetadataObject", {deviceInfo: {GAClientId: request.payload}});
		}
	}
);
