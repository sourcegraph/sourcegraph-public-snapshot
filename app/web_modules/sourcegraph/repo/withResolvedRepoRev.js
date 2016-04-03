// @flow weak

import React from "react";

import Container from "sourcegraph/Container";
import RepoStore from "sourcegraph/repo/RepoStore";
import "sourcegraph/repo/RepoBackend";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import Dispatcher from "sourcegraph/Dispatcher";
import {repoPath, repoRev, repoParam} from "sourcegraph/repo";

export default function withResolvedRepoRev(Component) {
	class WithResolvedRepoRev extends Container {
		static propTypes = {
			params: React.PropTypes.object.isRequired,
		};

		stores() {
			return [RepoStore];
		}

		reconcileState(state, props) {
			Object.assign(state, props);

			const repoSplat = repoParam(props.params.splat);
			state.repo = repoPath(repoSplat);
			state.rev = repoRev(repoSplat);

			state.repoObj = RepoStore.repos.get(state.repo);
			if (!state.rev) state.rev = state.repoObj && state.repoObj.DefaultBranch || null;

			state.isCloning = state.repoObj && state.repoObj.IsCloning || false;
		}

		onStateTransition(prevState, nextState) {
			if (prevState.repo !== nextState.repo) {
				Dispatcher.Backends.dispatch(new RepoActions.WantRepo(nextState.repo));
			}
		}

		render() {
			return <Component {...this.props} {...this.state} />;
		}
	}

	return WithResolvedRepoRev;
}
