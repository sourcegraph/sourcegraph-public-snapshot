import * as React from "react";
import * as Relay from "react-relay";
import {PageTitle} from "sourcegraph/components/PageTitle";
import {Repos} from "sourcegraph/user/settings/Repos";

interface Props {
	location: any;
}

type State = any;

export class UserSettingsReposMainComponent extends React.Component<Props & {root: GQL.IRoot}, {}> {
	render(): JSX.Element | null {
		const {root} = this.props;
		return (
			<div>
				<PageTitle title="Repositories"/>
				<Repos repos={root.remoteRepositories} location={this.props.location} />
			</div>
		);
	}
}

const UserSettingsReposMainContainer = Relay.createContainer(UserSettingsReposMainComponent, {
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
				}
			}
		`,
	},
});

export const UserSettingsReposMain = function(props: Props): JSX.Element {
	return <Relay.RootContainer
		Component={UserSettingsReposMainContainer}
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
