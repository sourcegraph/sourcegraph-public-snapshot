import { context } from "sourcegraph/app/context";
import { hubSpotEventNames } from "sourcegraph/util/constants/AnalyticsConstants";

interface HubSpotScript {
	push: ([]: any) => void;
}

class HubSpotWrapper {
	hubSpot: HubSpotScript | null;
	constructor() {
		if (global && global.window && global.window._hsq) {
			this.hubSpot = global.window._hsq;
		} else {
			console.error("Error loading HubSpot script. global.window._hsq not present.");
		}
	}

	logHubSpotEvent(eventLabel: string): void {
		if (!this.hubSpot) {
			return;
		}
		if (context.userAgentIsBot) {
			return;
		}
		if (!hubSpotEventNames.has(eventLabel)) {
			return;
		}
		this.hubSpot.push(["trackEvent", { id: eventLabel }]);
	}

	setHubSpotProperties(hubSpotAttributes: { email?: string, user_id?: string, fullname?: string, company?: string, location?: string, is_private_code_user?: string, emails?: string, authed_orgs_github?: string }): void {
		if (!this.hubSpot) {
			return;
		}
		this.hubSpot.push(["identify", hubSpotAttributes]);
	}

}

export const HubSpot = new HubSpotWrapper();
