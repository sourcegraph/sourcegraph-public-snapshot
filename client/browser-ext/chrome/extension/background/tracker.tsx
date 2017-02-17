import { TelligentWrapper } from "../../../app/tracking/TelligentWrapper";
import { Domain, getDomain } from "../../../app/utils";

let telligentWrapper: TelligentWrapper | null = null;

switch (getDomain(window.location)) {
	case Domain.GITHUB:
	// TODO(uforic): the app should probably be different here for domain Sourcegraph,
	// but this was existing behavior
	case Domain.SOURCEGRAPH:
		telligentWrapper = new TelligentWrapper("SourcegraphExtension", "BrowserExtension", true, true);
		break;
	case Domain.SGDEV_PHABRICATOR:
		telligentWrapper = new TelligentWrapper("SourcegraphExtension", "BrowserExtension", false, false);
		break;
	default:
		break;
}

/**
 * These messages come from the ExtensionEventLogger. This has to run in the background
 * because it requires access to cookies, and the foreground of Chrome extensions
 * don't have access to that.
 */
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
	if (!telligentWrapper) {
		return;
	}
	if (request.type === "trackEvent") {
		telligentWrapper.track(request.payload.eventAction, request.payload);
	} else if (request.type === "trackView") {
		telligentWrapper.track("view", request.payload);
	} else if (request.type === "setTrackerUserId") {
		telligentWrapper.setUserId(request.payload);
	} else if (request.type === "setTrackerDeviceId") {
		telligentWrapper.addStaticMetadataObject({ deviceInfo: { TelligentWebDeviceId: request.payload } });
	} else if (request.type === "setTrackerGAClientId") {
		telligentWrapper.addStaticMetadataObject({ deviceInfo: { GAClientId: request.payload } });
	}
});
