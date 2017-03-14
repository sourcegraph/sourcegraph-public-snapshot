import * as React from "react";

import { Org, OrgMember } from "sourcegraph/api";
import { context } from "sourcegraph/app/context";
import { RouterContext } from "sourcegraph/app/router";
import { GridCol, GridRow, Heading, TabItem, Tabs } from "sourcegraph/components";
import { PageTitle } from "sourcegraph/components/PageTitle";
import { colors, whitespace } from "sourcegraph/components/utils";
import { Container } from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as OrgActions from "sourcegraph/org/OrgActions";
import { OrgCard } from "sourcegraph/org/OrgCard";
import { OrgPanel } from "sourcegraph/org/OrgPanel";
import { OrgStore } from "sourcegraph/org/OrgStore";
import { Store } from "sourcegraph/Store";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";

interface State {
	orgs: Org[] | null;
	selectedOrg: Org | null;
	members: OrgMember[] | null;
}

export class OrgContainer extends Container<{}, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: RouterContext;
	state: State = {
		orgs: OrgStore.orgs || null,
		selectedOrg: null,
		members: OrgStore.members || null,
	};

	reconcileState(state: State): void {
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
		return <div
			style={{ marginTop: whitespace[8], padding: whitespace[8], textAlign: "center", maxWidth: 500, marginLeft: "auto", marginRight: "auto" }}>
			<Heading level={5}>
				<span>It looks like you're not a part of any organizations.</span>
			</Heading>
			<div style={{ color: colors.blueGray() }}>
				<span>
					Don't see the organization you were looking for? Your organization's GitHub permissions may restrict third-party applications.
					You can <a target="_blank" href="https://github.com/settings/connections/applications/8ac4b6c4d2e7b0721d68">request access</a>
					on GitHub, or contact us at <a href="mailto:hi@sourcegraph.com">hi@sourcegraph.com</a>.
			</span>;
			</div>
		</div>;
	}

	_onSelectOrg(org: Org): void {
		Events.Org_Selected.logEvent({ org_name: org.Login });
		this.setState(
			Object.assign({}, this.state, { selectedOrg: org })
		);
	}

	render(): JSX.Element {
		let mainPanel;
		if (!this.state.selectedOrg) {
			mainPanel = <div style={{ marginTop: whitespace[5], paddingTop: whitespace[3], paddingBottom: whitespace[3] }}>
				Select an organization to view and invite members.
			</div>;
		} else if (this.state.selectedOrg) {
			mainPanel = <OrgPanel org={this.state.selectedOrg} members={this.state.members} />;
		}
		return <div>
			<Heading level={5} style={{
				marginTop: whitespace[3],
				marginBottom: whitespace[3],
				marginLeft: whitespace[4],
				marginRight: whitespace[4],
			}}>Organization settings</Heading>
			<hr style={{ borderColor: colors.blueGrayL3(0.7) }} />
			<PageTitle title="Organization settings" />
			<div style={{ marginTop: whitespace[2] }}>
				{(!this._hasOrgs()) ? this._noRepoPanel() :
					<GridRow>
						<GridCol style={{ paddingTop: whitespace[4], paddingRight: whitespace[0] }} align="left" col={3} colSm={10}>
							<Tabs direction="vertical" style={{ borderLeft: "none" }}>
								{(this.state.orgs && this.state.orgs.length > 0) && this.state.orgs.map((org, i) =>
									<TabItem key={i} active={Boolean(this.state.selectedOrg && (this.state.selectedOrg.Login === org.Login))} direction="vertical">
										<a onClick={this._onSelectOrg.bind(this, org)}>
											<OrgCard org={org} />
										</a>
									</TabItem>
								)}
							</Tabs>
						</GridCol>
						<GridCol align="right" col={9} colSm={11}>{mainPanel}</GridCol>
					</GridRow>
				}
			</div>
		</div>;
	}
}
