class Logger {
	constructor() {
		if (process.env.NODE_ENV === "test") {
			return;
		}

		chrome.runtime.sendMessage({ type: "getIdentity" }, (identity) => {
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

	updatePropsForUser(identity?: any): void {
		if (identity && identity.userId) {
			chrome.runtime.sendMessage({ type: "setTrackerUserId", payload: identity.userId });
		}
		if (identity && identity.deviceId) {
			chrome.runtime.sendMessage({ type: "setTrackerDeviceId", payload: identity.deviceId });
		}
		if (identity && identity.gaClientId) {
			chrome.runtime.sendMessage({ type: "setTrackerGAClientId", payload: identity.gaClientId });
		}
	}

	_logToConsole(eventAction: string, object: any): void {
		if (window.localStorage["log_debug"]) {
			// tslint:disable-next-line
			console.debug("%cEVENT %s", "color: #aaa", eventAction, object);
		}
	}

	_defaultProperties(): Object {
		return {
			path_name: window.location.pathname,
			Platform: window.navigator.userAgent.indexOf("Firefox") !== -1 ? "FirefoxExtension" : "ChromeExtension",
		};
	}

	logEventForCategory(eventCategory: string, eventAction: string, eventLabel: string, eventProperties: Object = {}): void {
		if (process.env.NODE_ENV === "test") {
			return;
		}

		const decoratedEventProps = Object.assign({}, eventProperties, this._defaultProperties(),
			{
				eventLabel,
				eventCategory,
				eventAction,
			},
		);

		this._logToConsole(eventAction, decoratedEventProps);
		chrome.runtime.sendMessage({ type: "trackEvent", payload: decoratedEventProps });
	}

	// Use logViewEvent as the default way to log view events for Telligent and GA
	// location is the URL, page is the path.
	logViewEvent(title: string, page: string, eventProperties: Object): void {
		if (process.env.NODE_ENV === "test") {
			return;
		}

		const decoratedEventProps = Object.assign({}, eventProperties, this._defaultProperties(),
			{
				page_name: page,
				page_title: title,
			},
		);

		this._logToConsole(title, decoratedEventProps);
		chrome.runtime.sendMessage({ type: "trackView", payload: decoratedEventProps });
	}

}

export const EventLogger = new Logger();
