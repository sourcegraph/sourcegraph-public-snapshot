import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { EventLogger } from "sourcegraph/tracking/EventLogger";

export function successHandler(pageName: string): void {
	Events.ChromeExtension_Installed.logEvent({ page_name: pageName });
	EventLogger.setUserInstalledChromeExtension("true");
	// Syncs the our site analytics tracking with the chrome extension tracker.
	EventLogger.updateTrackerWithIdentificationProps();
}

export function failHandler(pageName: string): void {
	Events.ChromeExtensionInstall_Failed.logEvent({ page_name: pageName });
	EventLogger.setUserInstalledChromeExtension("false");
}

export function installChromeExtensionClicked(pageName: string): void {
	Events.ChromeExtensionCTA_Clicked.logEvent({ page_name: pageName });

	if (!!global.chrome) {
		Events.ChromeExtensionInstall_Started.logEvent({ page_name: pageName });
		global.chrome.webstore.install("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", successHandler.bind(null, pageName), failHandler.bind(null, pageName));
	} else {
		Events.ChromeExtensionStore_Redirected.logEvent({ page_name: pageName });
		window.open("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", "_newtab");
	}
}
