// HubSpot's frontend tracking script has proven to be extremely unreliable —
// a significant number of users recieve no tracking, properties, or events
// on an irreproducible basis.
//
// As a result, THE HUBSPOT SCRIPT SHOULD NOT BE USED FOR ANY CRITICAL DATA
// LOGGING. Any important contact properties or event tracking should be done
// through form submissions to /.api/submit-form, on the backend through the
// pkg/hubspot package, or through other means (such as during the analytics 
// pipeline's ETL jobs).
//
// This script should be used only for user activity logging or other non-
// essential purposes.

interface HubSpotScript {
	push: ([]: any) => void;
}

interface HubSpotUserAttributes {
	authed_orgs_github?: string | null;
	email?: string | null;
	installed_chrome_extension?: string | null;
	invited_by_user?: string | null;
	invited_to_org?: string | null;
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

	setHubSpotProperties(hubSpotAttributes: HubSpotUserAttributes): void {
		const hsq = this.getHubspot();
		if (!hsq) {
			return;
		}
		hsq.push(["identify", hubSpotAttributes]);
	}

}

export const hubSpot = new HubSpotWrapper();
