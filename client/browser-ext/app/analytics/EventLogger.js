import React from "react";

export class EventLogger {
	constructor() {
		if (process.env.NODE_ENV === "test") return;

		if (global.window && global.window.chrome && chrome) {
			chrome.runtime.sendMessage(null, {type: "getIdentity"}, {}, (identity) => {
				if (identity) {
					if (identity.userId) {
						chrome.runtime.sendMessage({ type: "setTrackerUserId", payload: identity.userId });
					}
					if (identity.deviceId) {
						chrome.runtime.sendMessage({ type: "setTrackerDeviceId", payload: identity.deviceId });
					}
				}
			});
		}
	}

	updatePropsForUser(identity) {
		if (identity) {
			if (identity.userId) {
				chrome.runtime.sendMessage({ type: "setTrackerUserId", payload: identity.userId });
			}
			if (identity.deviceId) {
				chrome.runtime.sendMessage({ type: "setTrackerDeviceId", payload: identity.deviceId });
			}
			if (identity.gaClientId) {
				chrome.runtime.sendMessage({ type: "setTrackerGAClientId", payload: identity.gaClientId});
			}
		}
	}

	_logToConsole(eventAction, object) {
		if (global.window && global.window.localStorage && global.window.localStorage["log_debug"]) {
			console.debug("%cEVENT %s", "color: #aaa", eventAction, object);
		}
	}

	_defaultProperties() {
		return {
			path_name: global.window && global.window.location && global.window.location.pathname ? global.window.location.pathname.slice(1) : "",
			Platform: window.navigator.userAgent.indexOf("Firefox") !== -1 ? "FirefoxExtension" : "ChromeExtension",
		};
	}

	logEventForCategory(eventCategory, eventAction, eventLabel, eventProperties = {}) {
		if (process.env.NODE_ENV === "test") return;

		const decoratedEventProps = Object.assign({}, eventProperties, this._defaultProperties(),
			{
				eventLabel,
				eventCategory,
				eventAction,
			},
		);

		this._logToConsole(eventAction, decoratedEventProps);
		chrome.runtime.sendMessage({ type: "trackEvent", payload: decoratedEventProps});
	}

	// Use logViewEvent as the default way to log view events for Telligent and GA
	// location is the URL, page is the path.
	logViewEvent(title, page, eventProperties) {
		if (process.env.NODE_ENV === "test") return;

		const decoratedEventProps = Object.assign({}, eventProperties, this._defaultProperties(),
			{
				page_name: page,
				page_title: title,
			},
		);

		this._logToConsole(title, decoratedEventProps);
		chrome.runtime.sendMessage({ type: "trackView", payload: decoratedEventProps});
	}

}

export default new EventLogger();
