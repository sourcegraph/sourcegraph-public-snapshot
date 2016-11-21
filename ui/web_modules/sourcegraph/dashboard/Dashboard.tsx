import * as React from "react";
import * as Relay from "react-relay";
import {PageTitle} from "sourcegraph/components/PageTitle";
import {Location} from "sourcegraph/Location";
import {Repos} from "sourcegraph/user/settings/Repos";

interface Props { location: Location; }

class DashboardComponent extends React.Component<Props & { root: GQL.IRoot }, {}> {
	render(): JSX.Element {
		return <div>
			<PageTitle title="My repositories" />
			<Repos repos={this.props.root.remoteRepositories} location={this.props.location} />
		</div>;
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
					httpCloneURL
					language
					fork
					mirror
					private
					createdAt
					pushedAt
					contributors {
						avatarURL
						login
						contributions
					}
				}
			}
		`,
	},
});

export const Dashboard = function(props: Props): JSX.Element {
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
