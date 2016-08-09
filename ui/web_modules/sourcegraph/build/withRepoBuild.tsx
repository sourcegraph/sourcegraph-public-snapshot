// tslint:disable

import * as React from "react";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as BuildActions from "sourcegraph/build/BuildActions";
import BuildStore from "sourcegraph/build/BuildStore";
import "sourcegraph/build/BuildBackend";

// withRepoBuild wraps Component and passes it a "build" property that holds
// the Build object for the most recently created build for the commit,
// if any.
export default function withRepoBuild(Component) {
	type Props = {
		repoID?: number,
		commitID?: string,
	};

	class WithRepoBuild extends Container<Props, any> {
		stores() { return [BuildStore]; }

		reconcileState(state, props) {
			Object.assign(state, props);
			const builds = state.repoID && state.commitID ? BuildStore.builds.listNewestByCommitID(state.repoID, state.commitID) : null;
			if (!builds) state.build = null;
			else if (builds && builds.length > 0) state.build = builds[0];
			else state.build = {Error: {response: {status: 404}}};
		}

		onStateTransition(prevState, nextState) {
			if (prevState.repoID !== nextState.repoID || prevState.commitID !== nextState.commitID || (!nextState.build && prevState.build !== nextState.build)) {
				if (!nextState.build && nextState.repoID) {
					Dispatcher.Backends.dispatch(new BuildActions.WantNewestBuildForCommit(nextState.repoID, nextState.commitID, false));
				}
			}
		}

		render(): JSX.Element | null {
			return <Component {...this.state} />;
		}
	}

	return WithRepoBuild;
}
