import { media, style } from "glamor";
import * as React from "react";
import * as Relay from "react-relay";
import { FlexContainer, PageTitle } from "sourcegraph/components";
import { colors, layout } from "sourcegraph/components/utils";
import { Repos } from "sourcegraph/dashboard/Repos";
import { TabBar } from "sourcegraph/dashboard/TabBar";
import { Location } from "sourcegraph/Location";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

export interface RepositoryTypes {
	"mine": any;
	"starred": any;
}

export type RepositoryTabs = "mine" | "starred";

interface Props { location: Location; }
interface State { active: RepositoryTabs; }

class DashboardComponent extends React.Component<Props & { root: GQL.IRoot }, State> {
	constructor(props: Props & { root: GQL.IRoot }) {
		super(props);
		this.state = { active: "mine" };
	}

	repositoryTypes: RepositoryTypes = {
		"mine": this.props.root.remoteRepositories,
		"starred": this.props.root.remoteStarredRepositories,
	};

	setActive(name: RepositoryTabs): void {
		this.setState({ active: name });
		AnalyticsConstants.Events.DashboardRepositoryTab_Clicked.logEvent({ name });
	}

	render(): JSX.Element {
		return <FlexContainer content="stretch" items="stretch" wrap={true} style={{
			alignSelf: "stretch",
			flex: "1 0",
		}}>

			<PageTitle title="Repositories" />

			<div
				{...style({ background: colors.coolGray2(), flex: "0 0 230px" }) }
				{...media(layout.breakpoints.sm, layout.flexItem.autoSize) }>
				<TabBar
					setActive={(name) => this.setActive(name)}
					active={this.state.active} />
			</div>

			<Repos
				type={this.state.active}
				repos={this.repositoryTypes[this.state.active]}
				location={this.props.location}
				style={{ flex: "1 1 500px", overflowY: "auto" }} />

		</FlexContainer>;
	}
}

const DashboardContainer = Relay.createContainer(DashboardComponent, {
	initialVariables: {
		repositories: null,
	},
	fragments: {
		root: () => Relay.QL`
			fragment on Root {
				remoteRepositories {
					uri
					name
					owner
					description
					language
				}
				remoteStarredRepositories {
					uri
					name
					owner
					description
					language
				}
			}
		`,
	},
});

export const Dashboard = function (props: Props): JSX.Element {
	return <Relay.RootContainer
		Component={DashboardContainer}
		route={{
			name: "Root",
			queries: {
				root: () => Relay.QL`
					query { root }
				`,
			},
			params: props,
		}}
		/>;
};
