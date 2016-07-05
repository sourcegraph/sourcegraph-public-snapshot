// @flow weak

import React from "react";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import "sourcegraph/repo/RepoBackend"; // for side effects
import RepoStore from "sourcegraph/repo/RepoStore";
import Repos from "sourcegraph/user/settings/Repos";
import * as RepoActions from "sourcegraph/repo/RepoActions";

export default class UserSettingsReposMain extends Container {
	static propTypes = {
		location: React.PropTypes.object.isRequired,
	};

	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
		user: React.PropTypes.object,
		signedIn: React.PropTypes.bool.isRequired,
		githubToken: React.PropTypes.object,
		eventLogger: React.PropTypes.object.isRequired,
		router: React.PropTypes.object,
	};

	reconcileState(state, props, context) {
		Object.assign(state, props);
		state.remoteRepos = RepoStore.remoteRepos.list();
		state.githubToken = context.githubToken;
		state.user = context.user;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.remoteRepos !== prevState.remoteRepos) {
			Dispatcher.Backends.dispatch(new RepoActions.WantRemoteRepos());
		}
	}

	stores() { return [RepoStore]; }

	render() {
		return <Repos repos={this.state.remoteRepos ? this.state.remoteRepos.Repos : null} location={this.props.location} />;
	}
}
