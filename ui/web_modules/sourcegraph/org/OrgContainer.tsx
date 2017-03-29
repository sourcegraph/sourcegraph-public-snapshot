import * as React from "react";

import { Org } from "sourcegraph/api";
import { context } from "sourcegraph/app/context";
import { RouterContext } from "sourcegraph/app/router";
import { FlexContainer, Heading, TabItem, Tabs } from "sourcegraph/components";
import { Spinner } from "sourcegraph/components/symbols";
import { colors, layout, typography, whitespace } from "sourcegraph/components/utils";
import { Container } from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as OrgActions from "sourcegraph/org/OrgActions";
import { OrgPanel } from "sourcegraph/org/OrgPanel";
import { OrgStore } from "sourcegraph/org/OrgStore";
import { Store } from "sourcegraph/Store";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";

interface State {
	selectedOrg: number;
}

export class OrgContainer extends Container<{}, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: RouterContext;

	state: State = { selectedOrg: 0 };

	stores(): Store<any>[] {
		return [OrgStore];
	}

	onStateTransition(prevState: State, nextState: State): void {
		if (!context.user || !context.hasOrganizationGitHubToken()) {
			return;
		}
		if (!OrgStore.orgs) {
			Dispatcher.Backends.dispatch(new OrgActions.WantOrgs(context.user.Login));
			return;
		}

		const org = OrgStore.orgs[nextState.selectedOrg];
		if (!org || org.Login == null) {
			return;
		}
		if (!OrgStore.members.get(org.Login)) {
			Dispatcher.Backends.dispatch(new OrgActions.WantOrgMembers(org.Login, String(org.ID)));
			return;
		}
		this.forceUpdate();
	}

	onSelectOrg(org: number): void {
		if (OrgStore.orgs) {
			Events.Org_Selected.logEvent({ org_name: OrgStore.orgs[org].Login });
			this.setState({ ...this.state, selectedOrg: org });
		}
	}

	renderOrgTabs(orgs: Org[]): JSX.Element[] {
		const { selectedOrg } = this.state;
		return orgs.map((org, i) => <TabItem key={i} active={selectedOrg === i} direction="vertical">
			<a onClick={() => this.onSelectOrg(i)} style={{ wordBreak: "break-word" }}>
				{org.Login}
			</a>
		</TabItem>);
	}

	render(): JSX.Element {
		const { selectedOrg } = this.state;
		const orgs = OrgStore.orgs;
		if (!orgs) {
			return <div style={{ textAlign: "center", padding: whitespace[4] }}><Spinner /></div>;
		}
		if (orgs.length === 0) {
			return <NoRepos />;
		}

		const members = OrgStore.members.get(orgs[selectedOrg].Login);
		const borderSx = `1px ${colors.blueGrayL2(0.5)} solid`;
		return <div style={{ margin: whitespace[4] }}>
			<FlexContainer items="start">
				<div style={{
					borderRadius: 3,
					border: borderSx,
					flex: "0 0 160px"
				}}>
					<Heading level={7} color="gray" style={{
						...typography.small,
						borderBottom: borderSx,
						paddingBottom: whitespace[2],
						paddingLeft: whitespace[3],
					}}>Organizations</Heading>
					<Tabs direction="vertical" style={{ ...typography.small, borderLeft: "none" }}>
						{this.renderOrgTabs(orgs)}
					</Tabs>
				</div>
				<div style={layout.flexItem.autoGrow}>
					<OrgPanel org={orgs[selectedOrg]} members={members} />
				</div>
			</FlexContainer>
		</div>;
	}
}

const NoRepos = () => <div style={{
	margin: "auto",
	maxWidth: 600,
	padding: whitespace[8],
	textAlign: "center",
}}>
	<Heading level={5}>We couldn't find any organizations</Heading>
	<div style={{ color: colors.blueGray() }}>
		<p>
			Not what you were expecting? Your organization's GitHub permissions may restrict third-party applications.
		</p>
		<p>
			You can <a target="_blank" href="https://github.com/settings/connections/applications/8ac4b6c4d2e7b0721d68">request access</a> on GitHub, or contact us at <a href="mailto:hi@sourcegraph.com">hi@sourcegraph.com</a>.
		</p>
	</div>
</div>;
