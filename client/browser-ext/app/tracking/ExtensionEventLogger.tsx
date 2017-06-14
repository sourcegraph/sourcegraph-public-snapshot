import { isE2ETest } from "../utils";
import { EventLogger } from "./EventLogger";

export class ExtensionEventLogger extends EventLogger {
	constructor() {
		super();
		this.updateIdentity();
	}

	updateIdentity(): void {
		if (isE2ETest()) {
			return;
		}
		chrome.runtime.sendMessage({ type: "getIdentity" }, this.updatePropsForUser);
	}

	updatePropsForUser(identity?: any): void {
		if (isE2ETest()) {
			return;
		}
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

	protected sendEvent(_: string, eventProps: any): void {
		chrome.runtime.sendMessage({ type: "trackEvent", payload: eventProps });
	}

}
