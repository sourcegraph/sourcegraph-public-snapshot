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
			rev: React.PropTypes.string,
			commitID: React.PropTypes.string,
			path: React.PropTypes.string,
		};

		stores() {
			return [BlobStore];
		}

		reconcileState(state, props) {
			Object.assign(state, props);

			state.anns = state.path && state.commitID ? BlobStore.annotations.get(state.repo, state.commitID, state.path, 0, 0) : null;
			const contentLenth = state.blob && !state.blob.Error ? state.blob.ContentsString.length : 0;
			state.skipAnns = contentLenth >= 40*2500; // ~ 2500 lines, avg. 40 chars per line
		}

		onStateTransition(prevState, nextState) {
			if (!nextState.anns && nextState.path && (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.commitID !== nextState.commitID || prevState.path !== nextState.path)) {
				if (nextState.commitID && !nextState.skipAnns) {
					// Require that the rev has been resolved to a commit ID to fetch,
					// so that we reuse that resolution on the client (which ensures
					// consistency and frees the server from performing repetitive
					// resolutions). Also require that the file isn't above line count
					// threshold for fetching annotations.
					Dispatcher.Backends.dispatch(new BlobActions.WantAnnotations(nextState.repo, nextState.commitID, nextState.path, 0, 0));
				}
			}
		}

		render() {
			return <Component {...this.props} {...this.state} />;
		}
	}

	return WithAnnotations;
}
