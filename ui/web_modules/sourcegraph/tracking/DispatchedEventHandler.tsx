import * as OrgActions from "sourcegraph/org/OrgActions";
import { Events, LogUnknownEvent } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { EventLogger } from "sourcegraph/tracking/EventLogger";

export function __onDispatch(action: any): void {
	switch (action.constructor) {
		case OrgActions.OrgsFetched:
			let orgNames: Array<string> = [];
			if (action.data) {
				for (let orgs of action.data) {
					orgNames.push(orgs.Login);
					if (orgs.Login === "sourcegraph" || orgs.Login === "sourcegraphtest") {
						EventLogger.setUserIsEmployee(true);
					}
				}
				EventLogger.setUserGitHubAuthedOrgs(orgNames);
				Events.AuthedOrgsGitHub_Fetched.logEvent({ "fetched_orgs_github": orgNames });
			}
			break;
		case OrgActions.OrgMembersFetched:
			if (action.data && action.orgName) {
				let orgName: string = action.orgName;
				let orgMemberNames: string[] = [];
				let orgMemberEmails: string[] = [];
				for (let member of action.data) {
					orgMemberNames.push(member.Login);
					orgMemberEmails.push(member.Email || "");
				}
				Events.AuthedOrgMembersGitHub_Fetched.logEvent({ "fetched_org_github": orgName, "fetched_org_member_names_github": orgMemberNames, "fetched_org_member_emails_github": orgMemberEmails });
			}
			break;
		default:
			// All dispatched actions to stores will automatically be tracked by the eventName
			// of the action (if set). Override this behavior by including another case above.
			if (action.eventObject) {
				action.eventObject.logEvent();
			} else if (action.eventName) {
				LogUnknownEvent(action.eventName);
			}
			break;
	}

	EventLogger.updateUser();
}
