import React from "react";

export class EventLogger {
    constructor() {
        if (process.env.NODE_ENV === "test") return;

        if (global.window && global.window.chrome && chrome) {
            chrome.runtime.sendMessage(null, {type: "getIdentity"}, {}, (identity) => {
                if (identity.userId) {
                    chrome.runtime.sendMessage({ type: "setTrackerUserId", payload: identity.userId });
                }
                if (identity.deviceId) {
                    chrome.runtime.sendMessage({ type: "setTrackerDeviceId", payload: identity.deviceId });
                }
            });
        }
    }

    updatePropsForUser(identity) {
        if (identity) {
            chrome.runtime.sendMessage({ type: "setTrackerUserId", payload: identity.userId });
            chrome.runtime.sendMessage({ type: "setTrackerDeviceId", payload: identity.deviceId });
            if (identity.gaClientId) {
                chrome.runtime.sendMessage({ type: "setTrackerGAClientId", payload: identity.gaClientId});
            }
        }
    }

	_decorateEventProperties(eventProperties) {
		return Object.assign({}, eventProperties, {path_name: global.window && global.window.location && global.window.location.pathname ? global.window.location.pathname.slice(1) : ""});
	}

	_logToConsole(eventAction, object) {
		if (global.window && global.window.localStorage && global.window.localStorage["log_debug"]) {
			console.debug("%cEVENT %s", "color: #aaa", eventAction, object);
		}
	}

    logEventForCategory(eventCategory, eventAction, eventLabel, eventProperties) {
        if (process.env.NODE_ENV === "test") return;

        eventProperties = eventProperties ? eventProperties : {};
        eventProperties["Platform"] = window.navigator.userAgent.indexOf("Firefox") !== -1 ? "FirefoxExtension" : "ChromeExtension";

        let props = Object.assign({}, eventProperties, { eventLabel, eventCategory, eventAction });
        const decoratedEventProps = this._decorateEventProperties(props);

        this._logToConsole(eventAction, decoratedEventProps);
        chrome.runtime.sendMessage({ type: "trackEvent", payload: decoratedEventProps});
    }
}

export default new EventLogger();
