// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import "sourcegraph/repo/RepoBackend"; // for side effects
import {RepoStore} from "sourcegraph/repo/RepoStore";
import {Repos} from "sourcegraph/user/settings/Repos";
import * as RepoActions from "sourcegraph/repo/RepoActions";

const reposQuerystring = "RemoteOnly=true";

type Props = {
	location: any,
};

export class UserSettingsReposMain extends Container<Props, any> {
	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
		user: React.PropTypes.object,
		signedIn: React.PropTypes.bool.isRequired,
		githubToken: React.PropTypes.object,
		eventLogger: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
	};

	reconcileState(state, props, context) {
		Object.assign(state, props);
		state.repos = RepoStore.repos.list(reposQuerystring);
		state.githubToken = context.githubToken;
		state.user = context.user;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.repos !== prevState.repos) {
			Dispatcher.Backends.dispatch(new RepoActions.WantRepos(reposQuerystring));
		}
	}

	stores() { return [RepoStore]; }

	render(): JSX.Element | null {
		return <Repos repos={this.state.repos ? this.state.repos.Repos : null} location={this.props.location} />;
	}
}
