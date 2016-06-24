import React from "react";

export class EventLogger {
	_amplitude: any = null;

	constructor() {
		if (process.env.NODE_ENV === "test") return;

		if (global.window && !this._amplitude) {
			this._amplitude = require("amplitude-js/amplitude");

			let apiKey = "34df64abd14304b388f884b46e0abbe2";
			if (process.env.NODE_ENV === "production") {
				apiKey = "e3c885c30d2c0c8bf33b1497b17806ba";
			}

			this._amplitude.init(apiKey, null, {
					includeReferrer: true,
					saveEvents: true,
					includeUtm: true,
					includeReferrer: true,
					savedMaxCount: 50,
			});
		}
	}

	logEvent(eventName, eventProperties) {
		if (process.env.NODE_ENV === "test") return;

		eventProperties = eventProperties ? eventProperties : {};
		// TODO(rothfels): this is broken, since Firefox has a polyfill that adds chrome
		// to the global scope.
		eventProperties["Platform"] = global.chrome ? "ChromeExtension" : "FirefoxExtension";
		this._amplitude.logEvent(eventName, eventProperties);
	}
}

export default new EventLogger();
