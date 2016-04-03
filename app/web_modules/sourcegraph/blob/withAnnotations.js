// @flow weak

import React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import BlobStore from "sourcegraph/blob/BlobStore";
import "sourcegraph/blob/BlobBackend";

// withAnnotations wraps Component and triggers a load of the annotations
// for the repo, rev, and path passed to it as properties.
export default function withAnnotations(Component) {
	class WithAnnotations extends Container {
		static propTypes = {
			repo: React.PropTypes.string.isRequired,
			rev: React.PropTypes.string.isRequired,
			path: React.PropTypes.string,
		};

		stores() {
			return [BlobStore];
		}

		reconcileState(state, props) {
			Object.assign(state, props);

			state.anns = state.path ? BlobStore.annotations.get(state.repo, state.rev, "", state.path, 0, 0) : null;
		}

		onStateTransition(prevState, nextState) {
			if (nextState.path && (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.path !== nextState.path)) {
				Dispatcher.Backends.dispatch(new BlobActions.WantAnnotations(nextState.repo, nextState.rev, "", nextState.path, 0, 0));
			}
		}

		render() {
			return <Component {...this.props} {...this.state} />;
		}
	}

	return WithAnnotations;
}
