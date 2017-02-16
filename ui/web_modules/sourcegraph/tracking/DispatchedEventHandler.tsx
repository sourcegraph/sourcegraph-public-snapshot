import * as OrgActions from "sourcegraph/org/OrgActions";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import { Events, LogUnknownEvent } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { EventLogger } from "sourcegraph/tracking/EventLogger";
import * as UserActions from "sourcegraph/user/UserActions";

export function __onDispatch(action: any): void {
	switch (action.constructor) {
		case RepoActions.ReposFetched:
			if (action.isUserRepos) {
				if (action.data.Repos) {
					let languages: Array<string> = [];
					let repos: Array<string> = [];
					let repoOwners: Array<string> = [];
					let repoNames: Array<string> = [];
					for (let repo of action.data.Repos) {
						languages.push(repo["Language"]);
						repoNames.push(repo["Name"]);
						repoOwners.push(repo["Owner"]);
						repos.push(` ${repo["Owner"]}/${repo["Name"]}`);
					}

					EventLogger.setUserGitHubAuthedLanguages(_dedupedArray(languages));
					EventLogger.setUserNumRepos(action.data.Repos.length);
					Events.RepositoryAuthedLanguagesGitHub_Fetched.logEvent({ "fetched_languages_github": _dedupedArray(languages) });
					Events.RepositoryAuthedReposGitHub_Fetched.logEvent({ "fetched_repo_names_github": _dedupedArray(repoNames), "fetched_repo_owners_github": _dedupedArray(repoOwners), "fetched_repos_github": _dedupedArray(repos) });
				}
			}
			break;
		case UserActions.BetaSubscriptionCompleted:
			if (action.eventObject) {
				action.eventObject.logEvent();
			}
			break;
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

function _dedupedArray(inputArray: Array<string>): Array<string> {
	return inputArray.filter(function (elem: string, index: number, self: any): any {
		return elem && (index === self.indexOf(elem));
	});
}
