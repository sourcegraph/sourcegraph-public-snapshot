import * as Raven from "raven-js";

if (process.env.NODE_ENV === "production") {
	const opt = {
		release: chrome.runtime.getManifest().version,
		tags: {
			platform: window.navigator.userAgent.indexOf("Firefox") !== -1 ? "firefox" : "chrome",
		},
		shouldSendCallback: (data) => data && data.extra && data.extra.extension,  // only log errors explicitly marked by extension
	};

	Raven.config("https://bf352e8f4a3541ee9ab9b8fce425b3b6@sentry.io/116423", opt).install();
}

export function setUser(id: string): void {
	if (process.env.NODE_ENV === "production") {
		Raven.setUserContext({id});
	}
}

export function logError(msg: string): void {
	const err = new Error(msg);
	logException(err);
}

export function logException(err: Error): void {
	console.error(err);
	if (process.env.NODE_ENV === "production") {
		Raven.captureException(err, {extra: {extension: true}});
	}
}
