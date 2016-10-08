import * as Dispatcher from "sourcegraph/Dispatcher";
import * as OrgActions from "sourcegraph/org/OrgActions";
import {OrgStore} from "sourcegraph/org/OrgStore";
import {checkStatus, defaultFetch} from "sourcegraph/util/xhr";

class OrganizationBackendClass {
	fetch: any;

	constructor() {
		this.fetch = defaultFetch;
	}

	__onDispatch(payload: any): void {
		if (payload instanceof OrgActions.WantOrgs) {
			const orgs = OrgStore.orgs;
			if (orgs === null) {
				this.fetch("/.api/orgs", {
						method: "POST",
						body: JSON.stringify({
							Username: payload.username,
						}),
					})
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => {
						Dispatcher.Stores.dispatch(new OrgActions.OrgsFetched(data.Orgs, payload.username));
					});
			}
		} else if (payload instanceof OrgActions.WantOrgMembers) {
			this.fetch("/.api/org-members", {
					method: "POST",
					body: JSON.stringify({
						OrgName: payload.orgName,
						OrgID: payload.orgID,
					}),
				})
				.then(checkStatus)
				.then((resp) => resp.json())
				.catch((err) => ({Error: err}))
				.then((data) => {
					Dispatcher.Stores.dispatch(new OrgActions.OrgMembersFetched(data.OrgMembers, payload.orgName));
				});
		} else if (payload instanceof OrgActions.SubmitOrgInvitation) {
			const action = payload;
			this.fetch(`/.api/org-invites`, {
				method: "POST",
				body: JSON.stringify({
					UserID: action.externalUserID,
					UserEmail: action.externalUserEmail,
					OrgName: action.externalOrgName,
					OrgID: action.externalOrgID,
				}),
			})
				.then(checkStatus)
				.then((resp) => resp.json())
				.catch((err) => ({Error: err}))
				.then(function(data: any): void {
					Dispatcher.Stores.dispatch(new OrgActions.OrgInvitationComplete(data));
					OrganizationBackend.__onDispatch(new OrgActions.WantOrgMembers(payload.externalOrgName, payload.externalOrgID));
				});
		}
	}
};

export const OrganizationBackend = new OrganizationBackendClass();
Dispatcher.Backends.register(OrganizationBackend.__onDispatch.bind(OrganizationBackend));
