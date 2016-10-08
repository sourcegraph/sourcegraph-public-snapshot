import { OrgMembersList, OrgsList } from "sourcegraph/api";

export type Action =
	WantOrgs | OrgsFetched |
	WantOrgMembers | OrgMembersFetched |
	SubmitOrgInvitation | OrgInvitationComplete;

export class WantOrgs {
	username: string;

	constructor(username: string) {
		this.username = username;
	}
}

export class OrgsFetched {
	data: OrgsList;
	username: string;

	constructor(data: OrgsList, username: string) {
		this.data = data;
		this.username = username;
	}
}

export class WantOrgMembers {
	orgName: string;
	orgID: string;

	constructor(orgName: string, orgID: string) {
		this.orgName = orgName;
		this.orgID = orgID;
	}
}

export class OrgMembersFetched {
	orgName: string;
	data: OrgMembersList;

	constructor(data: OrgMembersList, orgName: string) {
		this.orgName = orgName;
		this.data = data;
	}
}

export class SubmitOrgInvitation {
	externalUserID: string;
	externalUserEmail: string;
	externalOrgName: string;
	externalOrgID: string;

	constructor(externalUserID: string, externalUserEmail: string, externalOrgName: string, externalOrgID: string) {
		this.externalUserID = externalUserID;
		this.externalUserEmail = externalUserEmail;
		this.externalOrgName = externalOrgName;
		this.externalOrgID = externalOrgID;
	}
}

export class OrgInvitationComplete {
	orgName: string;

	constructor(orgName: string) {
		this.orgName = orgName;
	}
}
