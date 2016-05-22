// @flow weak

import React from "react";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as BuildActions from "sourcegraph/build/BuildActions";
import BuildStore from "sourcegraph/build/BuildStore";
import "sourcegraph/build/BuildBackend";

// withRepoBuild wraps Component and passes it a "build" property that holds
// the Build object for the most recently created build for the commit,
// if any.
export default function withRepoBuild(Component) {
	class WithRepoBuild extends Container {
		static propTypes = {
			repo: React.PropTypes.string.isRequired,
			commitID: React.PropTypes.string,
		};

		stores() { return [BuildStore]; }

		reconcileState(state, props) {
			Object.assign(state, props);
			const builds = state.commitID ? BuildStore.builds.listNewestByCommitID(state.repo, state.commitID) : null;
			if (!builds) state.build = null;
			else if (builds && builds.length > 0) state.build = builds[0];
			else state.build = {Error: {response: {status: 404}}};
		}

		onStateTransition(prevState, nextState) {
			if (prevState.repo !== nextState.repo || prevState.commitID !== nextState.commitID || (!nextState.build && prevState.build !== nextState.build)) {
				if (!nextState.build) {
					Dispatcher.Backends.dispatch(new BuildActions.WantNewestBuildForCommit(nextState.repo, nextState.commitID, false));
				}
			}
		}

		render() {
			return <Component {...this.state} />;
		}
	}

	return WithRepoBuild;
}
