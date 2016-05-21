// @flow weak

import React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import BlobStore from "sourcegraph/blob/BlobStore";
import "sourcegraph/blob/BlobBackend";
import {rel} from "sourcegraph/app/routePatterns";
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
		};

		stores() {
			return [BlobStore];
		}

		reconcileState(state, props) {
			Object.assign(state, props);
			state.path = props.route.path.startsWith(rel.blob) ? props.params.splat[1] : props.path;
			if (!state.path) state.path = null;

			// For defs, props.commitID is set to the resolved commit ID (if any);
			// for files, it is null, and the rev from the URL is all we can fetch by.
			state.fileRev = props.commitID || state.rev;
			state.blob = state.path ? BlobStore.files.get(state.repo, state.fileRev, state.path) : null;
			if (!state.commitID) state.commitID = state.blob && !state.blob.Error ? state.blob.CommitID : null;
		}

		onStateTransition(prevState, nextState) {
			// Handle change in params OR lost file contents (due to auth change, etc.).
			if (nextState.path && !nextState.blob && (prevState.repo !== nextState.repo || prevState.fileRev !== nextState.fileRev || prevState.path !== nextState.path || prevState.blob !== nextState.blob)) {
				Dispatcher.Backends.dispatch(new BlobActions.WantFile(nextState.repo, nextState.fileRev, nextState.path));
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
