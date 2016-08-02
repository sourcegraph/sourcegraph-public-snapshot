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
		if (login) _amplitude.getInstance().setUserId(login);
	}

	updateAmplitudePropsForUser(identity) {
		if (identity) {
			_amplitude.getInstance().setDeviceId(identity.deviceId);
			_amplitude.getInstance().setUserId(identity.userId);
		}
	}

	_isFirefox() {
		return window.navigator.userAgent.indexOf("Firefox") !== -1;
	}

	logEvent(eventName, eventProperties) {
		if (process.env.NODE_ENV === "test") return;

		eventProperties = eventProperties ? eventProperties : {};
		eventProperties["Platform"] = this._isFirefox() ? "FirefoxExtension" : "ChromeExtension";
		_amplitude.getInstance().logEvent(eventName, eventProperties);
	}
}

export default new EventLogger();
