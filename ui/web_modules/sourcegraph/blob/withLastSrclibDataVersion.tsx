// tslint:disable: typedef ordered-imports

import {TreeStore} from "sourcegraph/tree/TreeStore";
import "sourcegraph/tree/TreeBackend";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {Helper} from "sourcegraph/blob/BlobLoader";
import {rel} from "sourcegraph/app/routePatterns";

// withLastSrclibDataVersion sets the commitID property for all
// children to be that of the last srclib data version, not the
// originally passed in commit ID. This occurs only when loading
// files directly; it is not needed when the URL is to a def,
// since httpapi.serveDef performs the last-srclib-data-version
// resolution automatically.
//
// Assumes it runs AFTER withFileBlob (which sets state.path) but
// before withDefAndRefLocations (which uses the state.commitID that
// this sets).
export const withLastSrclibDataVersion = ({
	reconcileState(state, props) {
		// If a blob, then the path is statically known. Otherwise, reuse
		// the state.path set after the def loads (that is taken from def.File).
		state.path = props.route && props.route.path && props.route.path.startsWith(rel.blob) ? props.params.splat[1] : state.path;
		if (!state.path) {
			state.path = null;
		}

		// If we specify the path, then srclib-data-version resolution
		// is stricter: if the named file has changed since the last
		// build, resolution will fail. We only want this behavior when
		// the URL contains an explicit revision (such as a branch or commit).
		state.srclibDataVersionPath = state.path && props.rev ? state.path : null;
		state.srclibDataVersion = TreeStore.srclibDataVersions.get(props.repo, props.commitID, state.srclibDataVersionPath);

		// Set state.commitID to null until we know which commit ID to use (either
		// the srclib-last-version commit ID, or if there is no recent srclib data,
		// then the original commit ID). This avoids 2 network fetches and render
		// cycles where we first show the original then the srclib-resolved version.
		state.vcsCommitID = props.commitID;
		if (!state.srclibDataVersion) {
			state.commitID = null;
		} else if (state.srclibDataVersion.Error || !state.srclibDataVersion.CommitID) {
			state.commitID = props.commitID; // from URL
		} else {
			state.commitID = state.srclibDataVersion.CommitID;
		}
	},

	onStateTransition(prevState, nextState) {
		// Handle change in params OR lost def data (due to auth change, etc.).
		if (nextState.repo !== prevState.repo || nextState.rev !== prevState.rev || nextState.vcsCommitID !== prevState.vcsCommitID || nextState.srclibDataVersionPath !== prevState.srclibDataVersionPath || (!nextState.srclibDataVersion && nextState.srclibDataVersion !== prevState.srclibDataVersion)) {
			if (nextState.vcsCommitID) {
				Dispatcher.Backends.dispatch(new TreeActions.WantSrclibDataVersion(nextState.repo, nextState.vcsCommitID, nextState.srclibDataVersionPath));
			}
		}
	},

	store: TreeStore,
} as Helper);
