// tslint:disable-next-line
const telligent = require("telligent-tracker");
const telligentFunctionName = "telligent";

// Create the initializing function
window[telligentFunctionName] = function(): void {
	(window[telligentFunctionName].q = window[telligentFunctionName].q || []).push(arguments);
};

// Set up the initial queue, if it doesn't already exist
window[telligentFunctionName].q = new telligent.Telligent((window[telligentFunctionName].q || []), telligentFunctionName);

const t = (window as any).telligent;

// Must be called once upon initialization
t("newTracker", "SourcegraphExtensionTracker", "sourcegraph-logging.telligentdata.com", {
	encodeBase64: false,
	appId: "SourcegraphExtension",
	platform: "BrowserExtension",
	env: process.env.NODE_ENV,
	forceSecureTracker: true,
});

t("addStaticMetadata", "installed_chrome_extension", "true", "userInfo");

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
	if (request.type === "trackEvent") {
		t("track", request.payload.eventAction, request.payload);
	} else if (request.type === "trackView") {
		t("track", "view", request.payload);
	} else if (request.type === "setTrackerUserId") {
		t("setUserId", request.payload);
	} else if (request.type === "setTrackerDeviceId") {
		t("addStaticMetadataObject", {deviceInfo: {TelligentWebDeviceId: request.payload}});
	} else if (request.type === "setTrackerGAClientId") {
		t("addStaticMetadataObject", {deviceInfo: {GAClientId: request.payload}});
	}
});
