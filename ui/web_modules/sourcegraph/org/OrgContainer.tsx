import * as React from "react";
import Helmet from "react-helmet";
import {Org, OrgMember} from "sourcegraph/api";
import {context} from "sourcegraph/app/context";
import {GitHubAuthButton, GridCol, GridRow, Heading, Panel, TabItem, Tabs} from "sourcegraph/components";
import {colors} from "sourcegraph/components/utils";
import {whitespace} from "sourcegraph/components/utils/whitespace";
import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Location} from "sourcegraph/Location";
import * as OrgActions from "sourcegraph/org/OrgActions";
import {OrgCard} from "sourcegraph/org/OrgCard";
import {OrgPanel} from "sourcegraph/org/OrgPanel";
import {OrgStore} from "sourcegraph/org/OrgStore";
import {Store} from "sourcegraph/Store";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";
import {privateGitHubOAuthScopes} from "sourcegraph/util/urlTo";

interface Props {
	location: Location;
}

interface State {
	orgs: Org[] | null;
	selectedOrg: Org | null;
	members: OrgMember[] | null;
}

export class OrgContainer extends Container<Props, State> {
	constructor(props: Props) {
		super(props);
		this.state = {
			orgs: OrgStore.orgs || null,
			selectedOrg: null,
			members: OrgStore.members || null,
		};
	}

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);

		state.orgs = OrgStore.orgs;

		if (state.orgs) {
			if (state.orgs.length === 1) {
				state.selectedOrg = state.orgs[0];
			}
			if (state.selectedOrg) {
				state.members = OrgStore.members.get(state.selectedOrg.Login);
			}
		}
	}

	stores(): Store<any>[] {
		return [OrgStore];
	}

	onStateTransition(prevState: State, nextState: State): void {
		if (!context.user || !context.hasOrganizationGitHubToken()) {
			return;
		}

		if (!prevState.orgs) {
			Dispatcher.Backends.dispatch(new OrgActions.WantOrgs(context.user.Login));
		}

		let org = nextState.selectedOrg;
		if (!org || org.Login == null) {
			return;
		}

		if (org && org.Login && (!prevState.selectedOrg || prevState.selectedOrg.Login !== org.Login)) {
			Dispatcher.Backends.dispatch(new OrgActions.WantOrgMembers(org.Login, String(org.ID)));
		}
	}

	_hasOrgs(): boolean {
		return Boolean(this.state.orgs && this.state.orgs.length > 0);
	}

	_noRepoPanel(): JSX.Element {
		let msgHeader;
		let msgBody;

		if (context.hasPrivateGitHubToken()) {
			msgHeader = <span>It looks like you're not a part of any orgs</span>;
			msgBody = <span>
				If this doesn't seem right, try <a target="_blank" href="https://github.com/settings/connections/applications/8ac4b6c4d2e7b0721d68">verifying permissions</a> on GitHub.
			</span>;
		} else {
			msgHeader = <div>Browse your Org's private code on Sourcegraph</div>;
			msgBody = <div>
				Get inline annotations, jump to definition, and more for your company's private code.
				<div style={{marginTop: whitespace[4]}}>
					<GitHubAuthButton pageName={"ViewOrgs"} scopes={privateGitHubOAuthScopes} returnTo={"/settings"}>
						Add your orgs
					</GitHubAuthButton>
				</div>
			</div>;
		}

		return <Panel
			style={{marginTop: whitespace[5], padding: whitespace[5], textAlign: "center", maxWidth: 500, marginLeft: "auto", marginRight: "auto"}}>
			<Heading level={5}>
				{msgHeader}
			</Heading>
			<div style={{color: colors.coolGray3()}}>
				{msgBody}
			</div>
		</Panel>;
	}

	_onSelectOrg(org: Org): void {
		EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ORGS, AnalyticsConstants.ACTION_CLICK, "SelectedOrg", {org_name: org.Login});
		this.setState(
			Object.assign({}, this.state, {selectedOrg: org})
		);
	}

	render(): JSX.Element {
		let mainPanel;
		if (!this.state.selectedOrg) {
			mainPanel = <div style={{marginTop: whitespace[4], paddingTop: whitespace[3]}}>
				Choose an Org to the left to get started.
			</div>;
		} else if (this.state.selectedOrg) {
			mainPanel = <OrgPanel location={this.props.location} org={this.state.selectedOrg} members={this.state.members} />;
		} else {
			mainPanel = <div style={{marginTop: whitespace[4], paddingTop: whitespace[3]}}> Choose an Org to the left to get started.</div>;
		}

		return (
			<div>
				<Helmet title="Organizations" />
				<div style={{marginTop: whitespace[2]}}>
					{(!this._hasOrgs()) ? this._noRepoPanel() :
						<GridRow>
							<GridCol style={{paddingTop: whitespace[4], paddingRight: whitespace[0]}} align="left" col={3} colSm={10}>
								<Tabs direction="vertical">
									{(this.state.orgs && this.state.orgs.length > 0) && this.state.orgs.map((org, i) =>
										<TabItem key={i} active={Boolean(this.state.selectedOrg && (this.state.selectedOrg.Login === org.Login))}>
											<a onClick={this._onSelectOrg.bind(this, org)}>
												<OrgCard org={org}/>
											</a>
										</TabItem>
									)}
								</Tabs>
							</GridCol>
							<GridCol align="right" col={9} colSm={11}>{mainPanel}</GridCol>
						</GridRow>
					}
				</div>
			</div>
		);
	}
}
