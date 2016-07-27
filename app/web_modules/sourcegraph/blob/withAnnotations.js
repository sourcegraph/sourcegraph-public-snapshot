import * as React from "react";

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
			let contentLength = 0;
			if (state.blob && !state.blob.Error && state.blob.ContentsString) contentLength = state.blob.ContentsString.length;
			state.skipAnns = contentLength >= 60*10000; // ~ 10000 lines, avg. 60 chars per line
		}

		onStateTransition(prevState, nextState) {
			if (!nextState.anns && nextState.path && (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.commitID !== nextState.commitID || prevState.path !== nextState.path || prevState.blob !== nextState.blob)) {
				if (nextState.commitID && !nextState.skipAnns && nextState.blob) {
					// Require that the rev has been resolved to a commit ID to fetch,
					// so that we reuse that resolution on the client (which ensures
					// consistency and frees the server from performing repetitive
					// resolutions). Also require that the file isn't above line count
					// threshold for fetching annotations.
					//
					// Also wait until the file has fetched the blob (or gotten an error)
					// because the server usually includes the annotations in the blob
					// response. This means we rarely will have to actually reach this
					// line and trigger another network fetch to get the annotations.
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
