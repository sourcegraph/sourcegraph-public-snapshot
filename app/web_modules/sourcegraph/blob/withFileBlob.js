import * as React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import BlobStore, {keyForFile} from "sourcegraph/blob/BlobStore";
import "sourcegraph/blob/BlobBackend";
import {urlToTree} from "sourcegraph/tree/routes";

// withFileBlob wraps Component and passes it a "blob" property containing
// the blob fetched from the server. The path is taken from props or parsed from
// the URL (in that order).
//
// If the path refers to a tree, a redirect occurs.
export default function withFileBlob(Component) {
	class WithFileBlob extends Container {
		static contextTypes = {
			router: React.PropTypes.object.isRequired,
		};

		static propTypes = {
			repo: React.PropTypes.string.isRequired,
			rev: React.PropTypes.string,
			commitID: React.PropTypes.string,
			params: React.PropTypes.object.isRequired,
			path: React.PropTypes.string,
		};

		stores() {
			return [BlobStore];
		}

		reconcileState(state, props) {
			Object.assign(state, props);
			state.blob = state.path && state.commitID ? (BlobStore.files[keyForFile(state.repo, state.commitID, state.path)] || null) : null;
		}

		onStateTransition(prevState, nextState) {
			// Handle change in params OR lost file contents (due to auth change, etc.).
			if (nextState.path && nextState.commitID && !nextState.blob && (prevState.repo !== nextState.repo || prevState.commitID !== nextState.commitID || prevState.path !== nextState.path || prevState.blob !== nextState.blob)) {
				Dispatcher.Backends.dispatch(new BlobActions.WantFile(nextState.repo, nextState.commitID, nextState.path));
			}

			if (nextState.blob && prevState.blob !== nextState.blob) {
				// If the entry is a tree (not a file), redirect to the "/tree/" URL.
				// Run in setTimeout because it warns otherwise.
				if (nextState.blob.Entries) {
					setTimeout(() => {
						this.context.router.replace(urlToTree(nextState.repo, nextState.rev, nextState.path));
					});
				}
			}
		}

		render() {
			return <Component {...this.props} {...this.state} />;
		}
	}

	return WithFileBlob;
}
