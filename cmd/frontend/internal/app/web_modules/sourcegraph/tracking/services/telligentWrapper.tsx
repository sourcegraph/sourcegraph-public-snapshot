import { isOnPremInstance, sourcegraphContext } from "sourcegraph/util/sourcegraphContext";

declare global {
	interface Window {
		telligent: (...args: any[]) => void | null;
	}
}

class TelligentWrapper {
	private telligent: (...args: any[]) => void | null;
	private DEFAULT_ENV: string = "development";
	private PROD_ENV: string = "production";
	private DEFAULT_APP_ID: string = "UnknownApp";

	constructor() {
		if (window && window.telligent) {
			this.telligent = window.telligent;
		} else {
			return;
		}
		if (sourcegraphContext.version !== "dev" && sourcegraphContext.trackingAppID) {
			this.initialize(sourcegraphContext.trackingAppID, this.PROD_ENV);
		} else {
			this.initialize(this.DEFAULT_APP_ID, this.DEFAULT_ENV);
		}
	}

	isTelligentLoaded(): boolean {
		return Boolean(this.telligent);
	}

	setUserId(login: string): void {
		if (!this.telligent) {
			return;
		}
		this.telligent("setUserId", login);
	}

	addStaticMetadataObject(metadata: any): void {
		if (!this.telligent) {
			return;
		}
		this.telligent("addStaticMetadataObject", metadata);
	}

	setUserProperty(property: string, value: any): void {
		if (!this.telligent) {
			return;
		}
		this.telligent("addStaticMetadata", property, value);
	}

	track(eventAction: string, eventProps: any): void {
		if (!this.telligent) {
			return;
		}
		// TODO(Dan): validate umami user id props
		this.telligent("track", eventAction, eventProps);
	}

	private initialize(appId: string, env: string): void {
		if (!this.telligent) {
			return;
		}
		let telligentUrl = "sourcegraph-logging.telligentdata.com";
		// for an on-prem trial, we want to send information directly telligent.
		// for clients like umami, we use a bi-logger
		if (isOnPremInstance(sourcegraphContext.authEnabled) && sourcegraphContext.trackingAppID === "UmamiWeb") {
			telligentUrl = `${window.location.host}`.concat("/.bi-logger");
		}
		this.telligent("newTracker", "sg", telligentUrl, {
			appId: appId,
			platform: "Web",
			encodeBase64: false,
			env: env,
			configUseCookies: true,
			useCookies: true,
			cookieDomain: "sourcegraph.com",
			metadata: {
				gaCookies: true,
				performanceTiming: true,
				augurIdentityLite: true,
				webPage: true,
			},
		});
	}

	/**
	 * Function to extract the Telligent user ID from the first-party cookie set by the Telligent JavaScript Tracker
	 * @return string or bool The ID string if the cookie exists or null if the cookie has not been set yet
	 */
	getTelligentDuid(): string | null {
		const cookieProps = this.inspectTelligentCookie();
		return cookieProps ? cookieProps[0] : null;
	}

	/**
	 * Function to extract the Telligent session ID from the first-party cookie set by the Telligent JavaScript Tracker
	 * @return string or bool The session ID string if the cookie exists or null if the cookie has not been set yet
	 */
	getTelligentSessionId(): string | null {
		const cookieProps = this.inspectTelligentCookie();
		return cookieProps ? cookieProps[5] : null;
	}

	private inspectTelligentCookie(): string[] | null {
		const cookieName = "_te_";
		const matcher = new RegExp(cookieName + "id\\.[a-f0-9]+=([^;]+);?");
		const match = window.document.cookie.match(matcher);
		if (match && match[1]) {
			return match[1].split(".");
		} else {
			return null;
		}
	}
}

export const telligent = new TelligentWrapper();
