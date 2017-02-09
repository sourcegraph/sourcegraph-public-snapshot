import { context } from "sourcegraph/app/context";
import { hubSpotEventNames } from "sourcegraph/util/constants/AnalyticsConstants";

interface HubSpotScript {
	push: ([]: any) => void;
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

	setHubSpotProperties(hubSpotAttributes: { email?: string, user_id?: string, fullname?: string, company?: string, location?: string, is_private_code_user?: string, emails?: string, authed_orgs_github?: string }): void {
		const hsq = this.getHubspot();
		if (!hsq) {
			return;
		}
		hsq.push(["identify", hubSpotAttributes]);
	}

}

export const hubSpot = new HubSpotWrapper();
