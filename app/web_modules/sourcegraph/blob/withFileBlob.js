// @flow weak

import React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import BlobStore from "sourcegraph/blob/BlobStore";
import "sourcegraph/blob/BlobBackend";
import {rel} from "sourcegraph/app/routePatterns";

// withFileBlob wraps Component and passes it a "blob" property containing
// the blob fetched from the server. The path is taken from props or parsed from
// the URL (in that order).
//
// If the path refers to a tree, a redirect occurs. (TODO: not yet implemented.)
export default function withFileBlob(Component) {
	class WithFileBlob extends Container {
		static propTypes = {
			repo: React.PropTypes.string.isRequired,
			rev: React.PropTypes.string.isRequired,
			params: React.PropTypes.object.isRequired,
		};

		stores() {
			return [BlobStore];
		}

		reconcileState(state, props) {
			Object.assign(state, props);
			state.path = props.route.path.startsWith(rel.blob) ? props.params.splat[1] : props.path;
			if (!state.path) state.path = null;
			state.blob = state.path ? BlobStore.files.get(state.repo, state.rev, state.path) : null;
		}

		onStateTransition(prevState, nextState) {
			if (nextState.path && (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.path !== nextState.path)) {
				Dispatcher.Backends.dispatch(new BlobActions.WantFile(nextState.repo, nextState.rev, nextState.path));
			}
		}

		render() {
			return <Component {...this.props} {...this.state} />;
		}
	}

	return WithFileBlob;
}
