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

	updateAmplitudePropsForUser(identity) {
		if (identity) {
			_amplitude.getInstance().setDeviceId(identity.deviceId);
			_amplitude.getInstance().setUserId(identity.userId);
		}
	}

	logEvent(eventName, eventProperties) {
		if (process.env.NODE_ENV === "test") return;

		eventProperties = eventProperties ? eventProperties : {};
		// TODO(rothfels): this is broken, since Firefox has a polyfill that adds chrome
		// to the global scope.
		eventProperties["Platform"] = global.chrome ? "ChromeExtension" : "FirefoxExtension";
		_amplitude.getInstance().logEvent(eventName, eventProperties);
	}
}

export default new EventLogger();
