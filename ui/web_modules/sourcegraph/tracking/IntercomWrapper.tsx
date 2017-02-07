import { context } from "sourcegraph/app/context";

class IntercomWrapper {
	private intercom: (command: string, eventName?: any, settings?: any) => void | null;
	private intercomSettings: Object | null;
	constructor() {
		if (global && global.window && global.window.Intercom) {
			this.intercom = global.window.Intercom;
		} else {
			console.error("Error loading intercom script. global.window.Intercom not present.");
		}

		if (global && global.window && global.window.intercomSettings) {
			this.intercomSettings = global.window.intercomSettings;
		} else {
			console.error("Error loading intercom settings. global.window.intercomSettings not present.");
		}
	}

	boot(isOnPrem: boolean, trackingAppId: string | null): void {
		if (!this.intercom || !this.intercomSettings) {
			return;
		}
		this.intercom("boot", this.intercomSettings);
		this.setIntercomProperty("is_on_prem", isOnPrem);
		this.setIntercomProperty("tracking_app_id", trackingAppId);
	}

	logIntercomEvent(eventName: string, eventProperties: any): void {
		if (!this.intercom) {
			return;
		}
		if (context.userAgentIsBot) {
			return;
		}
		this.intercom("trackEvent", eventName, eventProperties);
	}

	setIntercomProperty(propertyId: string, value: any): void {
		if (!this.intercom || !this.intercomSettings) {
			return;
		}
		this.intercomSettings[propertyId] = value;
	}

	shutdown(): void {
		if (!this.intercom) {
			return;
		}
		this.intercom("shutdown");
	}

}

export const Intercom = new IntercomWrapper();
