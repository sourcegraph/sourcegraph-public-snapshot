import { context } from "sourcegraph/app/context";
import { hubSpotEventNames } from "sourcegraph/tracking/constants/AnalyticsConstants";

interface HubSpotScript {
	push: ([]: any) => void;
}

interface HubSpotUserAttributes {
	email?: string | null;
	user_id?: string | null;
	fullname?: string | null;
	company?: string | null;
	location?: string | null;
	is_private_code_user?: string | null;
	emails?: string | null;
	authed_orgs_github?: string | null;
	registered_at?: string | null;
}

class HubSpotWrapper {

	private getHubspot(): HubSpotScript | null {
		if (global && global.window && global.window._hsq) {
			return global.window._hsq;
		}
		return null;
	}

	logHubSpotEvent(eventLabel: string): void {
		const hsq = this.getHubspot();
		if (!hsq) {
			return;
		}
		if (context.userAgentIsBot) {
			return;
		}
		if (!hubSpotEventNames.has(eventLabel)) {
			return;
		}
		hsq.push(["trackEvent", { id: eventLabel }]);
	}

	setHubSpotProperties(hubSpotAttributes: HubSpotUserAttributes): void {
		const hsq = this.getHubspot();
		if (!hsq) {
			return;
		}
		hsq.push(["identify", hubSpotAttributes]);
	}

}

export const hubSpot = new HubSpotWrapper();
