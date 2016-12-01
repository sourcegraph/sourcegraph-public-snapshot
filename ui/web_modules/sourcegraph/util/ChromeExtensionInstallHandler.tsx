import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { EventLogger } from "sourcegraph/util/EventLogger";

export function successHandler(): void {
	AnalyticsConstants.Events.ChromeExtension_Installed.logEvent({ page_name: "MasterPlanPage" });
	EventLogger.setUserProperty("installed_chrome_extension", "true");
	// Syncs the our site analytics tracking with the chrome extension tracker.
	EventLogger.updateTrackerWithIdentificationProps();
}

export function failHandler(): void {
	AnalyticsConstants.Events.ChromeExtensionInstall_Failed.logEvent({ page_name: "MasterPlanPage" });
	EventLogger.setUserProperty("installed_chrome_extension", "false");
}

export function installChromeExtensionClicked(): void {
	AnalyticsConstants.Events.ChromeExtensionCTA_Clicked.logEvent({ page_name: "MasterPlanPage" });

	if (!!global.chrome) {
		AnalyticsConstants.Events.ChromeExtensionInstall_Started.logEvent({ page_name: "MasterPlanPage" });
		global.chrome.webstore.install("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", successHandler.bind(this), failHandler.bind(this));
	} else {
		AnalyticsConstants.Events.ChromeExtensionStore_Redirected.logEvent({ page_name: "MasterPlanPage" });
		window.open("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", "_newtab");
	}
}
