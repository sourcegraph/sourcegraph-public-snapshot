import * as React from "react";
import {context} from "sourcegraph/app/context";
import {GridCol, GridRow, Heading, Panel, TabItem, Tabs} from "sourcegraph/components";
import {whitespace} from "sourcegraph/components/utils/whitespace";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Location} from "sourcegraph/Location";
import * as OrgActions from "sourcegraph/org/OrgActions";
import {OrgContainer} from "sourcegraph/org/OrgContainer";
import {UserSettingsReposMain} from "sourcegraph/user/settings/UserSettingsReposMain";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";

interface Props {
	location: Location;
}

interface State {
	activePanel: number;
}

export class SettingsMain extends React.Component<Props, State>  {

	constructor() {
		super();
		this.state = {
			activePanel: 0,
		};
	}

	componentDidMount(): void {
		if (context.user) {
			Dispatcher.Backends.dispatch(new OrgActions.WantOrgs(context.user.Login));
		}
	}

	setActivePanel(panelIndex: number, e: React.MouseEvent<{}>): void {
		if (panelIndex === 0) {
			EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_REPOSITORY, AnalyticsConstants.ACTION_CLICK, "ToggleRepoView");
		} else {
			EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_REPOSITORY, AnalyticsConstants.ACTION_CLICK, "ToggleOrgView");
		}
		this.setState({activePanel: panelIndex});
		e.preventDefault();
	}

	render(): JSX.Element {

		const {location} = this.props;

		const panels = [
			<UserSettingsReposMain location={location} />,
			<OrgContainer location={location} />,
		];

		return <div>
			<GridRow style={{marginTop: whitespace[5], maxWidth: 1080, marginLeft: "auto", marginRight: "auto"}}>
				<GridCol col={2} colSm={11} align="left">
					<Heading level={7} ml={3} pl={1} mb={4} color="gray">Settings</Heading>
					<Tabs direction="vertical">
						<TabItem active={this.state.activePanel === 0}>
							<a onClick={this.setActivePanel.bind(this, 0)}>Repositories</a>
						</TabItem>
						<TabItem active={this.state.activePanel === 1}>
							<a onClick={this.setActivePanel.bind(this, 1)}>Organizations</a>
						</TabItem>
					</Tabs>
				</GridCol>
				<GridCol col={10} colSm={12} align="left">
					<Panel style={{padding: whitespace[0]}} hoverLevel="low" hover={false}>
						{panels[this.state.activePanel]}
					</Panel>
				</GridCol>
			</GridRow>
		</div>;
	}
}
