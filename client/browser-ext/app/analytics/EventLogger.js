import React from "react";

const _amplitude = require("amplitude-js/amplitude");

export class EventLogger {

	constructor() {
		if (process.env.NODE_ENV === "test") return;

		if (global.window) {
			let apiKey = "34df64abd14304b388f884b46e0abbe2";
			if (process.env.NODE_ENV === "production") {
				apiKey = "e3c885c30d2c0c8bf33b1497b17806ba";
			}

			_amplitude.getInstance().init(apiKey, null, {
					includeReferrer: true,
					saveEvents: true,
					includeUtm: true,
					includeReferrer: true,
					savedMaxCount: 50,
			});
		}
	}

	setUserLogin(login) {
		if (login) {
			chrome.runtime.sendMessage({type: "setTrackerUserId", payload: login});
			_amplitude.getInstance().setUserId(login);
		}
	}

	updatePropsForUser(identity) {
		if (identity) {
			chrome.runtime.sendMessage({type: "setTrackerUserId", payload: identity.userId});
			chrome.runtime.sendMessage({type: "setTrackerDeviceId", payload: identity.deviceId});
			_amplitude.getInstance().setDeviceId(identity.deviceId);
			_amplitude.getInstance().setUserId(identity.userId);
		}
	}

	logEventForCategory(eventCategory: string, eventAction: string, eventLabel: string, eventProperties?: any) {
		if (process.env.NODE_ENV === "test") return;
		
		eventProperties = eventProperties ? eventProperties : {};
		eventProperties["Platform"] = window.navigator.userAgent.indexOf("Firefox") !== -1 ? "FirefoxExtension" : "ChromeExtension";

		chrome.runtime.sendMessage({type: "trackEvent", payload: Object.assign({}, eventProperties, {eventLabel, eventCategory, eventAction})});
		_amplitude.getInstance().logEvent(eventLabel, Object.assign({}, eventProperties, {eventLabel, eventCategory, eventAction}));
	}
}

export default new EventLogger();
