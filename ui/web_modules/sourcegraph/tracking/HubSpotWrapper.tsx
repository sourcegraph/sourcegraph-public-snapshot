import { context } from "sourcegraph/app/context";

// Set of all Sourcegraph events (specifically, eventLabels) that should be sent to HubSpot.
const hubSpotEventNames = new Set(["SignupCompleted"]);

interface HubSpotScript {
	push: ([]: any) => void;
}

interface HubSpotUserAttributes {
	authed_orgs_github?: string | null;
	company?: string | null;
	email?: string | null;
	emails?: string | null;
	fullname?: string | null;
	github_name?: string | null;
	github_company?: string | null;
	github_link?: string | null;
	installed_chrome_extension?: string | null;
	invited_by_user?: string | null;
	invited_to_org?: string | null;
	is_private_code_user?: string | null;
	location?: string | null;
	looker_link?: string | null;
	plan?: string;
	plan_orgs?: string;
	registered_at?: string | null;
	user_id?: string | null;
	viewed_pricing_page?: string | null;
}

class HubSpotWrapper {

	// getHubspot either gets or creates a new HubSpot event stack
	// Per HubSpot API docs, if the _hsq Array hasn't been created because the
	// external script hasn't loaded yet, we can create an empty Array, which will
	// be flushed/recorded when loading is complete
	// https://knowledge.hubspot.com/events-user-guide-v2/using-custom-events
	private getHubspot(): HubSpotScript | null {
		if (global && global.window && global.window._hsq) {
			return global.window._hsq;
		}
		if (global && global.window) {
			global.window._hsq = [];
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
