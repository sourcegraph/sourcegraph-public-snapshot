import {Location} from "history";
import * as React from "react";
import {InjectedRouter} from "react-router";
import {Org, OrgMember} from "sourcegraph/api";
import {context} from "sourcegraph/app/context";
import {Loader} from "sourcegraph/components";
import {setLocationModalState} from "sourcegraph/components/Modal";
import {whitespace} from "sourcegraph/components/utils/whitespace";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as OrgActions from "sourcegraph/org/OrgActions";
import {OrgInviteModal} from "sourcegraph/org/OrgInviteModal";
import {OrgMembersTable} from "sourcegraph/org/OrgMembersTable";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";

interface Props {
	org: Org;
	members: OrgMember[];
	location: Location;
}

export class OrgPanel extends React.Component<Props, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: InjectedRouter };

	_invitedUser(member: OrgMember): void {
		if (member.Email != null && context.user != null && this.props.org.Login) {
			Dispatcher.Backends.dispatch(new OrgActions.SubmitOrgInvitation(member.Login || "", member.Email, this.props.org.Login, String(this.props.org.ID)));
			EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ORGS, AnalyticsConstants.ACTION_SUCCESS, "InviteUser", {org_name: this.props.org.Login, num_invites: 1});
		} else {
			EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ORGS, AnalyticsConstants.ACTION_TOGGLE, "ToggleManualInviteModal", {org_name: this.props.org.Login});
			setLocationModalState(this.context.router, this.props.location, "orgInvite", true);
		}
	}

	_onInviteUser(invites: Array<Object>): void {
		if (this.props.org && this.props.org.Login && context.user) {
			EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ORGS, AnalyticsConstants.ACTION_SUCCESS, "InviteUser", {org_name: this.props.org.Login, num_invites: invites.length});
			for (let i = 0; i < invites.length; i++) {
				let invite = invites[i];
				let member = invite["member"];
				Dispatcher.Backends.dispatch(new OrgActions.SubmitOrgInvitation(member["Login"] || "", invite["email"], this.props.org.Login, String(this.props.org.ID)));
			}

			setLocationModalState(this.context.router, this.props.location, "orgInvite", false);
		}
	}

	_orgMembersList(members: OrgMember[]): JSX.Element | null {
		if (!members) {
			return <div>
				<p>Fetching organization members</p>
				<span><Loader/></span>
			</div>;
		}

		if (members.length === 0) {
			return <div>
				<p>Looks like your organization is empty. Invite some of your users to join!</p>
			</div>;
		}

		return <OrgMembersTable inviteClicked={this._invitedUser.bind(this)} members={members} />;
	}

	render(): JSX.Element | null {
		let {members} = this.props;
		if (members === null) {
			return null;
		}

		let inviteMembers = members.filter((member: OrgMember): boolean => {
			return member.CanInvite === true;
		});

		return <div>
				<OrgInviteModal onInvite={this._onInviteUser.bind(this)} members={inviteMembers} location={this.props.location}/>
				<div style={{padding: whitespace[4]}}>{this._orgMembersList(members)}</div>
			</div>;
	}
}
