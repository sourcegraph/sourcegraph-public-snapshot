import React from "react";

export class EventLogger {

    updatePropsForUser(identity) {
        if (identity) {
            chrome.runtime.sendMessage({ type: "setTrackerUserId", payload: identity.userId });
            chrome.runtime.sendMessage({ type: "setTrackerDeviceId", payload: identity.deviceId });
            if (identity.gaClientId) {
                chrome.runtime.sendMessage({ type: "setTrackerGAClientId", payload: identity.gaClientId});
            }
        }
    }

	_decorateEventProperties(eventProperties: any) {
		return Object.assign({}, eventProperties, {path_name: global.window && global.window.location && global.window.location.pathname ? global.window.location.pathname.slice(1) : ""});
	}

	_logToConsole(eventAction: string, object?: any) {
		if (global.window && global.window.localStorage && global.window.localStorage["log_debug"]) {
			console.debug("%cEVENT %s", "color: #aaa", eventAction, object);
		}
	}

    logEventForCategory(eventCategory: string, eventAction: string, eventLabel: string, eventProperties ? : any) {
        if (process.env.NODE_ENV === "test") return;

        eventProperties = eventProperties ? eventProperties : {};
        eventProperties["Platform"] = window.navigator.userAgent.indexOf("Firefox") !== -1 ? "FirefoxExtension" : "ChromeExtension";

        const eventCtxt = this._decorateEventProperties(eventProperties, { eventLabel, eventCategory, eventAction });

        this._logToConsole(eventAction, Object.assign(eventCtxt));
        chrome.runtime.sendMessage({ type: "trackEvent", payload: Object.assign(eventCtxt) });
    }
}

export default new EventLogger();
