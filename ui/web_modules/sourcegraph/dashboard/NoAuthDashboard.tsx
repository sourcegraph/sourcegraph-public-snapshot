import { media, style } from "glamor";
import * as React from "react";
import * as Relay from "react-relay";
import { FlexContainer, PageTitle } from "sourcegraph/components";
import { colors, layout } from "sourcegraph/components/utils";
import { Props } from "sourcegraph/dashboard";
import { HelpBar } from "sourcegraph/dashboard/HelpBar";
import { Repos } from "sourcegraph/dashboard/Repos";

class NoAuthDashboardComponent extends React.Component<Props & { root: GQL.IRoot }, {}> {
	constructor(props: Props & { root: GQL.IRoot }) {
		super(props);
	}

	render(): JSX.Element {
		return <FlexContainer content="stretch" items="stretch" wrap={true} style={{
			alignSelf: "stretch",
			flex: "1 0",
		}}>

			<PageTitle title="Repositories" />

			<div
				{...style({ backgroundColor: colors.blueGrayD1(), flex: "0 0 230px" }) }
				{...media(layout.breakpoints.sm, layout.flexItem.autoSize) }>
				<HelpBar />
			</div>

			<Repos
				repos={this.props.root.repositories}
				location={this.props.location}
				style={{ flex: "1 1 500px", overflowY: "auto" }} />

		</FlexContainer>;
	}
}

export const NoAuthDashboard = function (props: Props): JSX.Element {
	return <Relay.RootContainer
		Component={NoAuthDashboardContainer}
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

const NoAuthDashboardContainer = Relay.createContainer(NoAuthDashboardComponent, {
	initialVariables: {
		repositories: null,
	},
	fragments: {
		root: () => Relay.QL`
			fragment on Root {
				repositories {
					uri
					language
				}
			}
		`,
	},
});
